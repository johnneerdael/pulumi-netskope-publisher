#[allow(
    clippy::doc_lazy_continuation,
    clippy::tabs_in_doc_comments,
    clippy::should_implement_trait
)]
pub mod kubernetes_publisher {
    #[derive(pulumi_gestalt_rust::__private::bon::Builder)]
    #[builder(finish_fn = build_struct)]
    #[allow(dead_code)]
    pub struct KubernetesPublisherArgs {
        #[builder(into, default)]
        pub api_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub auth_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub bearer_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub chart_repository: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub chart_values: pulumi_gestalt_rust::Input<
            Option<std::collections::HashMap<String, String>>,
        >,
        #[builder(into, default)]
        pub chart_version: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub enrollment_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub hpa_enabled: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub hpa_max_replicas: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub hpa_min_replicas: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub image_repository: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub image_tag: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub name_prefix: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub names: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into, default)]
        pub namespace: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub oauth2: pulumi_gestalt_rust::Input<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        #[builder(into, default)]
        pub registrations: pulumi_gestalt_rust::Input<
            Option<
                std::collections::HashMap<
                    String,
                    super::types::provider::PublisherRegistrationInput,
                >,
            >,
        >,
        #[builder(into, default)]
        pub replicas: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub tags: pulumi_gestalt_rust::Input<
            Option<std::collections::HashMap<String, String>>,
        >,
        #[builder(into, default)]
        pub tenant_url: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub wizard_path: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub workload_type: pulumi_gestalt_rust::Input<Option<String>>,
    }
    #[allow(dead_code)]
    pub struct KubernetesPublisherResult {
        /// Pulumi ID is the provider-assigned unique ID for this managed resource.
        /// It is set during deployments and may be missing (unknown) during planning phases.
        pub id: pulumi_gestalt_rust::Output<String>,
        /// Pulumi URN is the stable logical identity of this resource in the Pulumi stack.
        pub urn: pulumi_gestalt_rust::Output<String>,
        pub api_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub auth_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub bearer_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub chart_repository: pulumi_gestalt_rust::Output<Option<String>>,
        pub chart_values: pulumi_gestalt_rust::Output<
            Option<std::collections::HashMap<String, String>>,
        >,
        pub chart_version: pulumi_gestalt_rust::Output<Option<String>>,
        pub enrollment_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub helm_release_names: pulumi_gestalt_rust::Output<Vec<String>>,
        pub hpa_enabled: pulumi_gestalt_rust::Output<Option<bool>>,
        pub hpa_max_replicas: pulumi_gestalt_rust::Output<Option<i32>>,
        pub hpa_min_replicas: pulumi_gestalt_rust::Output<Option<i32>>,
        pub image_repository: pulumi_gestalt_rust::Output<Option<String>>,
        pub image_tag: pulumi_gestalt_rust::Output<Option<String>>,
        pub name_prefix: pulumi_gestalt_rust::Output<Option<String>>,
        pub names: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub namespace: pulumi_gestalt_rust::Output<Option<String>>,
        pub oauth2: pulumi_gestalt_rust::Output<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        pub publisher_names: pulumi_gestalt_rust::Output<Vec<String>>,
        pub publishers: pulumi_gestalt_rust::Output<
            std::collections::HashMap<String, String>,
        >,
        pub registrations: pulumi_gestalt_rust::Output<
            Option<
                std::collections::HashMap<
                    String,
                    super::types::provider::PublisherRegistrationInput,
                >,
            >,
        >,
        pub replicas: pulumi_gestalt_rust::Output<Option<i32>>,
        pub tags: pulumi_gestalt_rust::Output<
            Option<std::collections::HashMap<String, String>>,
        >,
        pub tenant_url: pulumi_gestalt_rust::Output<Option<String>>,
        pub wizard_path: pulumi_gestalt_rust::Output<Option<String>>,
        pub workload_type: pulumi_gestalt_rust::Output<Option<String>>,
    }
    ///
    /// Registers a new resource with the given unique name and arguments
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: KubernetesPublisherArgs,
    ) -> KubernetesPublisherResult {
        __create(ctx, name, args, None)
    }
    ///
    /// Same as `create`, but with additional generic options that control the behavior of the resource registration.
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create_with_options(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: KubernetesPublisherArgs,
        options: pulumi_gestalt_rust::CustomResourceOptions,
    ) -> KubernetesPublisherResult {
        __create(ctx, name, args, Some(options))
    }
    #[allow(non_snake_case, unused_imports, dead_code)]
    fn __create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: KubernetesPublisherArgs,
        options: Option<pulumi_gestalt_rust::CustomResourceOptions>,
    ) -> KubernetesPublisherResult {
        let api_token_binding = args.api_token.get_output(ctx);
        let auth_mode_binding = args.auth_mode.get_output(ctx);
        let bearer_token_binding = args.bearer_token.get_output(ctx);
        let chart_repository_binding = args.chart_repository.get_output(ctx);
        let chart_values_binding = args.chart_values.get_output(ctx);
        let chart_version_binding = args.chart_version.get_output(ctx);
        let enrollment_mode_binding = args.enrollment_mode.get_output(ctx);
        let hpa_enabled_binding = args.hpa_enabled.get_output(ctx);
        let hpa_max_replicas_binding = args.hpa_max_replicas.get_output(ctx);
        let hpa_min_replicas_binding = args.hpa_min_replicas.get_output(ctx);
        let image_repository_binding = args.image_repository.get_output(ctx);
        let image_tag_binding = args.image_tag.get_output(ctx);
        let name_prefix_binding = args.name_prefix.get_output(ctx);
        let names_binding = args.names.get_output(ctx);
        let namespace_binding = args.namespace.get_output(ctx);
        let oauth2_binding = args.oauth2.get_output(ctx);
        let registrations_binding = args.registrations.get_output(ctx);
        let replicas_binding = args.replicas.get_output(ctx);
        let tags_binding = args.tags.get_output(ctx);
        let tenant_url_binding = args.tenant_url.get_output(ctx);
        let wizard_path_binding = args.wizard_path.get_output(ctx);
        let workload_type_binding = args.workload_type.get_output(ctx);
        let request = pulumi_gestalt_rust::RegisterResourceRequest {
            type_: "netskope-publisher:index:KubernetesPublisher".into(),
            name: name.to_string(),
            version: super::get_version(),
            object: &[
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "apiToken".into(),
                    value: &api_token_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "authMode".into(),
                    value: &auth_mode_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "bearerToken".into(),
                    value: &bearer_token_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "chartRepository".into(),
                    value: &chart_repository_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "chartValues".into(),
                    value: &chart_values_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "chartVersion".into(),
                    value: &chart_version_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "enrollmentMode".into(),
                    value: &enrollment_mode_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "hpaEnabled".into(),
                    value: &hpa_enabled_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "hpaMaxReplicas".into(),
                    value: &hpa_max_replicas_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "hpaMinReplicas".into(),
                    value: &hpa_min_replicas_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "imageRepository".into(),
                    value: &image_repository_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "imageTag".into(),
                    value: &image_tag_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "namePrefix".into(),
                    value: &name_prefix_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "names".into(),
                    value: &names_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "namespace".into(),
                    value: &namespace_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "oauth2".into(),
                    value: &oauth2_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "registrations".into(),
                    value: &registrations_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "replicas".into(),
                    value: &replicas_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "tags".into(),
                    value: &tags_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "tenantUrl".into(),
                    value: &tenant_url_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "wizardPath".into(),
                    value: &wizard_path_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "workloadType".into(),
                    value: &workload_type_binding.drop_type(),
                },
            ],
            options,
        };
        let o = ctx.register_resource(request);
        KubernetesPublisherResult {
            id: o.get_id(),
            urn: o.get_urn(),
            api_token: o.get_field("apiToken"),
            auth_mode: o.get_field("authMode"),
            bearer_token: o.get_field("bearerToken"),
            chart_repository: o.get_field("chartRepository"),
            chart_values: o.get_field("chartValues"),
            chart_version: o.get_field("chartVersion"),
            enrollment_mode: o.get_field("enrollmentMode"),
            helm_release_names: o.get_field("helmReleaseNames"),
            hpa_enabled: o.get_field("hpaEnabled"),
            hpa_max_replicas: o.get_field("hpaMaxReplicas"),
            hpa_min_replicas: o.get_field("hpaMinReplicas"),
            image_repository: o.get_field("imageRepository"),
            image_tag: o.get_field("imageTag"),
            name_prefix: o.get_field("namePrefix"),
            names: o.get_field("names"),
            namespace: o.get_field("namespace"),
            oauth2: o.get_field("oauth2"),
            publisher_names: o.get_field("publisherNames"),
            publishers: o.get_field("publishers"),
            registrations: o.get_field("registrations"),
            replicas: o.get_field("replicas"),
            tags: o.get_field("tags"),
            tenant_url: o.get_field("tenantUrl"),
            wizard_path: o.get_field("wizardPath"),
            workload_type: o.get_field("workloadType"),
        }
    }
}
