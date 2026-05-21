import * as k8s from "@pulumi/kubernetes";
import * as pulumi from "@pulumi/pulumi";
import { validateComponentArgs } from "./providerValidation";
import {
  createRegistrations,
  resolvePublisherNames,
  requireManagedRegistrationInputs,
} from "./componentCore";
import {
  KubernetesPublisherArgs,
  KubernetesPublisherOutput,
} from "./types";

const defaultChartRepository = "oci://ghcr.io/johnneerdael/charts";
const defaultChartVersion = "~> 1.4";

export class KubernetesPublisher extends pulumi.ComponentResource {
  public readonly publisherNames: pulumi.Output<string[]>;
  public readonly helmReleaseNames: pulumi.Output<string[]>;
  public readonly publishers: pulumi.Output<Record<string, KubernetesPublisherOutput>>;

  constructor(name: string, args: KubernetesPublisherArgs, opts?: pulumi.ComponentResourceOptions) {
    super("netskope-publisher:index:KubernetesPublisher", name, {}, opts);
    validateComponentArgs("KubernetesPublisher", args);

    const parentOpts = { parent: this };
    const publisherNames = resolvePublisherNames(args);
    const namespaceName = args.namespace ?? "netskope";
    const enrollmentMode = args.enrollmentMode ?? "token";

    this.publisherNames = pulumi.output(publisherNames);

    const namespace = new k8s.core.v1.Namespace(`${name}-namespace`, {
      metadata: {
        name: namespaceName,
      },
    }, parentOpts);

    const namespaceOutput = pulumi.output(namespace.metadata).apply((metadata) => metadata.name ?? namespaceName);
    const releaseNames = pulumi.output(enrollmentMode).apply((mode) =>
      mode === "api" ? ["npa-publisher"] : publisherNames,
    );
    this.helmReleaseNames = releaseNames;

    const commonValues = pulumi.all([
      args.workloadType ?? "daemonset",
      args.hpaEnabled ?? false,
      args.hpaMinReplicas ?? 2,
      args.hpaMaxReplicas ?? 6,
      args.tags ?? {},
      args.imageRepository,
      args.imageTag,
    ]).apply(([workloadType, hpaEnabled, hpaMinReplicas, hpaMaxReplicas, tags, imageRepository, imageTag]) => {
      const image = {
        ...(imageRepository === undefined ? {} : { repository: imageRepository }),
        ...(imageTag === undefined ? {} : { tag: imageTag }),
      };

      return {
        workload: {
          type: workloadType,
        },
        hpa: {
          enabled: hpaEnabled && workloadType === "statefulset",
          minReplicas: hpaMinReplicas,
          maxReplicas: hpaMaxReplicas,
        },
        commonLabels: tags,
        ...(Object.keys(image).length === 0 ? {} : { image }),
      };
    });

    const releases: Record<string, k8s.helm.v3.Release> = {};
    const publisherOutputs: Record<string, pulumi.Output<KubernetesPublisherOutput>> = {};

    if (enrollmentMode === "api") {
      const required = requireManagedRegistrationInputs(args);
      const apiAuthMode = required.authMode ?? "token";
      const oauth2 = pulumi.output(required.oauth2!);
      const apiSecret = apiAuthMode === "oauth2"
        ? new k8s.core.v1.Secret(`${name}-api-oauth`, {
          metadata: {
            name: "npa-api-oauth",
            namespace: namespaceOutput,
          },
          stringData: {
            "client-id": oauth2.apply((values) => values.clientId),
            "client-secret": oauth2.apply((values) => values.clientSecret),
          },
          type: "Opaque",
        }, { ...parentOpts, dependsOn: [namespace] })
        : new k8s.core.v1.Secret(`${name}-api-token`, {
          metadata: {
            name: "npa-api-token",
            namespace: namespaceOutput,
          },
          stringData: {
            "api-token": required.bearerToken ?? required.apiToken!,
          },
          type: "Opaque",
        }, { ...parentOpts, dependsOn: [namespace] });

      const release = createHelmRelease(name, "npa-publisher", namespaceOutput, args, pulumi.all([commonValues, args.chartValues ?? {}]).apply(([values, chartValues]) => ({
        ...values,
        enrollment: {
          mode: "api",
          api: {
            baseUrl: required.tenantUrl,
            authMode: apiAuthMode,
            ...(apiAuthMode === "oauth2" ? {
              oauth2: {
                tokenUrl: oauth2.apply((values) => values.tokenUrl),
                existingSecret: "npa-api-oauth",
                clientIdKey: "client-id",
                clientSecretKey: "client-secret",
                scope: oauth2.apply((values) => values.scope ?? ""),
              },
            } : {
              existingSecret: "npa-api-token",
              tokenKey: "api-token",
            }),
            cleanupOnDelete: false,
          },
        },
        ...chartValues,
      })), [apiSecret], parentOpts);
      releases["npa-publisher"] = release;

      publisherOutputs["npa-publisher"] = pulumi.all([namespaceOutput, release.status]).apply(([namespace, status]) => ({
        publisherId: undefined,
        registrationToken: undefined,
        helmReleaseName: "npa-publisher",
        namespace,
        status: status?.status ?? "",
        vmId: undefined,
        privateIp: undefined,
        publicIp: undefined,
      }));
    } else {
      const registrations = createRegistrations(name, publisherNames, args, parentOpts);

      for (const publisherName of publisherNames) {
        const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
        const tokenSecret = new k8s.core.v1.Secret(`${name}-${publisherName}-registration-token`, {
          metadata: {
            name: `${publisherName}-registration-token`,
            namespace: namespaceOutput,
          },
          stringData: {
            token: registration.registrationToken,
          },
          type: "Opaque",
        }, { ...parentOpts, dependsOn: [namespace] });

        const release = createHelmRelease(name, publisherName, namespaceOutput, args, pulumi.all([commonValues, args.chartValues ?? {}]).apply(([values, chartValues]) => ({
          ...values,
          enrollment: {
            mode: "token",
            commonName: publisherName,
          },
          registrationToken: {
            existingSecret: `${publisherName}-registration-token`,
            existingSecretKey: "token",
          },
          ...chartValues,
        })), [tokenSecret], parentOpts);
        releases[publisherName] = release;

        publisherOutputs[publisherName] = pulumi.all([registration, namespaceOutput, release.status]).apply(([record, namespace, status]) => ({
          publisherId: record.publisherId,
          registrationToken: record.registrationToken,
          helmReleaseName: publisherName,
          namespace,
          status: status?.status ?? "",
          vmId: undefined,
          privateIp: undefined,
          publicIp: undefined,
        }));
      }
    }

    this.publishers = pulumi.secret(pulumi.all(publisherOutputs));
    this.registerOutputs({
      publisherNames: this.publisherNames,
      helmReleaseNames: this.helmReleaseNames,
      publishers: this.publishers,
    });
  }
}

function createHelmRelease(
  componentName: string,
  releaseName: string,
  namespace: pulumi.Input<string>,
  args: KubernetesPublisherArgs,
  values: pulumi.Input<Record<string, any>>,
  dependsOn: pulumi.Resource[],
  opts: pulumi.ComponentResourceOptions,
): k8s.helm.v3.Release {
  return new k8s.helm.v3.Release(`${componentName}-${releaseName}`, {
    name: releaseName,
    namespace,
    chart: "kubernetes-netskope-publisher",
    version: args.chartVersion ?? defaultChartVersion,
    repositoryOpts: {
      repo: args.chartRepository ?? defaultChartRepository,
    },
    createNamespace: false,
    atomic: true,
    skipAwait: false,
    timeout: 300,
    values,
  }, { ...opts, dependsOn });
}
