import * as pulumi from "@pulumi/pulumi";
import { renderUserDataBase64 } from "./cloudInit";
import { createPublisherOutput, createRegistrations, resolvePublisherNames } from "./componentCore";
import { CommonPublisherArgs, PublisherOutput } from "./types";

export interface VmPublisherRuntime {
  parent: pulumi.ComponentResource;
  componentName: string;
  args: CommonPublisherArgs;
  forceBootstrap?: boolean;
  defaultNonat?: boolean;
}

export interface VmPublisherBuildInput {
  publisherName: string;
  registration: pulumi.Output<{
    publisherId: number;
    registrationToken: string;
    existedBefore?: boolean;
  }>;
  userDataBase64: pulumi.Output<string>;
  userData: pulumi.Output<string>;
}

export interface VmPublisherBuildResult {
  vmId: pulumi.Input<string>;
  privateIp: pulumi.Input<string>;
  publicIp?: pulumi.Input<string>;
}

export function createVmPublishers(
  runtime: VmPublisherRuntime,
  build: (input: VmPublisherBuildInput) => VmPublisherBuildResult,
): {
  publisherNames: pulumi.Output<string[]>;
  publishers: pulumi.Output<Record<string, PublisherOutput>>;
} {
  const parentOpts = { parent: runtime.parent };
  const publisherNames = resolvePublisherNames(runtime.args);
  const registrations = createRegistrations(runtime.componentName, publisherNames, runtime.args, parentOpts);
  const publisherOutputs: Record<string, pulumi.Output<PublisherOutput>> = {};

  for (const publisherName of publisherNames) {
    const registration = registrations.apply((allRegistrations) => allRegistrations[publisherName]);
    const userDataBase64 = pulumi.all({
      registration,
      wizardPath: runtime.args.wizardPath,
      bootstrap: runtime.forceBootstrap ? true : runtime.args.bootstrap,
      bootstrapUrl: runtime.args.bootstrapUrl,
      nonat: runtime.args.nonat ?? runtime.defaultNonat ?? false,
      installUser: runtime.args.installUser,
      installUserPassword: runtime.args.installUserPassword,
      installUserPasswordIsHash: runtime.args.installUserPasswordIsHash,
      installUserSshAuthorizedKeys: runtime.args.installUserSshAuthorizedKeys,
      deleteDefaultUser: runtime.args.deleteDefaultUser,
      guestNetworkInterface: runtime.args.guestNetworkInterface,
    }).apply((options: any) =>
      renderUserDataBase64({
        publisherName,
        registrationToken: options.registration.registrationToken,
        wizardPath: options.wizardPath,
        bootstrap: options.bootstrap,
        bootstrapUrl: options.bootstrapUrl,
        nonat: options.nonat,
        installUser: options.installUser,
        installUserPassword: options.installUserPassword,
        installUserPasswordIsHash: options.installUserPasswordIsHash,
        installUserSshAuthorizedKeys: options.installUserSshAuthorizedKeys,
        deleteDefaultUser: options.deleteDefaultUser,
        guestNetworkInterface: options.guestNetworkInterface,
      }),
    );
    const userData = userDataBase64.apply((value) => Buffer.from(value, "base64").toString("utf8"));
    const vm = build({ publisherName, registration, userDataBase64, userData });

    publisherOutputs[publisherName] = createPublisherOutput({
      registration,
      vmId: vm.vmId,
      privateIp: vm.privateIp,
      publicIp: vm.publicIp,
    });
  }

  return {
    publisherNames: pulumi.output(publisherNames),
    publishers: pulumi.secret(pulumi.all(publisherOutputs)),
  };
}
