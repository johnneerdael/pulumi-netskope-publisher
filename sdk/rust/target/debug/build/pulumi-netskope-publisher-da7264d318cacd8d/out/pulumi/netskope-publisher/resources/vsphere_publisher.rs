#[allow(
    clippy::doc_lazy_continuation,
    clippy::tabs_in_doc_comments,
    clippy::should_implement_trait
)]
pub mod vsphere_publisher {
    #[derive(pulumi_gestalt_rust::__private::bon::Builder)]
    #[builder(finish_fn = build_struct)]
    #[allow(dead_code)]
    pub struct VspherePublisherArgs {
        #[builder(into, default)]
        pub api_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub auth_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub bearer_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub cluster: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into)]
        pub datacenter: pulumi_gestalt_rust::Input<String>,
        #[builder(into)]
        pub datastore: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub folder: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub host: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub memory: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub name_prefix: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub names: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into)]
        pub network_name: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub num_cpus: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub oauth2: pulumi_gestalt_rust::Input<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        #[builder(into, default)]
        pub placement_labels: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
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
        #[builder(into)]
        pub template_name: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub tenant_url: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub wizard_path: pulumi_gestalt_rust::Input<Option<String>>,
    }
    #[allow(dead_code)]
    pub struct VspherePublisherResult {
        /// Pulumi ID is the provider-assigned unique ID for this managed resource.
        /// It is set during deployments and may be missing (unknown) during planning phases.
        pub id: pulumi_gestalt_rust::Output<String>,
        /// Pulumi URN is the stable logical identity of this resource in the Pulumi stack.
        pub urn: pulumi_gestalt_rust::Output<String>,
        pub api_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub auth_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub bearer_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub cluster: pulumi_gestalt_rust::Output<Option<String>>,
        pub datacenter: pulumi_gestalt_rust::Output<String>,
        pub datastore: pulumi_gestalt_rust::Output<String>,
        pub folder: pulumi_gestalt_rust::Output<Option<String>>,
        pub host: pulumi_gestalt_rust::Output<Option<String>>,
        pub memory: pulumi_gestalt_rust::Output<Option<i32>>,
        pub name_prefix: pulumi_gestalt_rust::Output<Option<String>>,
        pub names: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub network_name: pulumi_gestalt_rust::Output<String>,
        pub num_cpus: pulumi_gestalt_rust::Output<Option<i32>>,
        pub oauth2: pulumi_gestalt_rust::Output<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        pub placement_labels: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
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
        pub template_name: pulumi_gestalt_rust::Output<String>,
        pub tenant_url: pulumi_gestalt_rust::Output<Option<String>>,
        pub wizard_path: pulumi_gestalt_rust::Output<Option<String>>,
    }
    ///
    /// Registers a new resource with the given unique name and arguments
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: VspherePublisherArgs,
    ) -> VspherePublisherResult {
        __create(ctx, name, args, None)
    }
    ///
    /// Same as `create`, but with additional generic options that control the behavior of the resource registration.
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create_with_options(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: VspherePublisherArgs,
        options: pulumi_gestalt_rust::CustomResourceOptions,
    ) -> VspherePublisherResult {
        __create(ctx, name, args, Some(options))
    }
    #[allow(non_snake_case, unused_imports, dead_code)]
    fn __create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: VspherePublisherArgs,
        options: Option<pulumi_gestalt_rust::CustomResourceOptions>,
    ) -> VspherePublisherResult {
        let api_token_binding = args.api_token.get_output(ctx);
        let auth_mode_binding = args.auth_mode.get_output(ctx);
        let bearer_token_binding = args.bearer_token.get_output(ctx);
        let cluster_binding = args.cluster.get_output(ctx);
        let datacenter_binding = args.datacenter.get_output(ctx);
        let datastore_binding = args.datastore.get_output(ctx);
        let folder_binding = args.folder.get_output(ctx);
        let host_binding = args.host.get_output(ctx);
        let memory_binding = args.memory.get_output(ctx);
        let name_prefix_binding = args.name_prefix.get_output(ctx);
        let names_binding = args.names.get_output(ctx);
        let network_name_binding = args.network_name.get_output(ctx);
        let num_cpus_binding = args.num_cpus.get_output(ctx);
        let oauth2_binding = args.oauth2.get_output(ctx);
        let placement_labels_binding = args.placement_labels.get_output(ctx);
        let registrations_binding = args.registrations.get_output(ctx);
        let replicas_binding = args.replicas.get_output(ctx);
        let tags_binding = args.tags.get_output(ctx);
        let template_name_binding = args.template_name.get_output(ctx);
        let tenant_url_binding = args.tenant_url.get_output(ctx);
        let wizard_path_binding = args.wizard_path.get_output(ctx);
        let request = pulumi_gestalt_rust::RegisterResourceRequest {
            type_: "netskope-publisher:index:VspherePublisher".into(),
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
                    name: "cluster".into(),
                    value: &cluster_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "datacenter".into(),
                    value: &datacenter_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "datastore".into(),
                    value: &datastore_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "folder".into(),
                    value: &folder_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "host".into(),
                    value: &host_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "memory".into(),
                    value: &memory_binding.drop_type(),
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
                    name: "networkName".into(),
                    value: &network_name_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "numCpus".into(),
                    value: &num_cpus_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "oauth2".into(),
                    value: &oauth2_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "placementLabels".into(),
                    value: &placement_labels_binding.drop_type(),
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
                    name: "templateName".into(),
                    value: &template_name_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "tenantUrl".into(),
                    value: &tenant_url_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "wizardPath".into(),
                    value: &wizard_path_binding.drop_type(),
                },
            ],
            options,
        };
        let o = ctx.register_resource(request);
        VspherePublisherResult {
            id: o.get_id(),
            urn: o.get_urn(),
            api_token: o.get_field("apiToken"),
            auth_mode: o.get_field("authMode"),
            bearer_token: o.get_field("bearerToken"),
            cluster: o.get_field("cluster"),
            datacenter: o.get_field("datacenter"),
            datastore: o.get_field("datastore"),
            folder: o.get_field("folder"),
            host: o.get_field("host"),
            memory: o.get_field("memory"),
            name_prefix: o.get_field("namePrefix"),
            names: o.get_field("names"),
            network_name: o.get_field("networkName"),
            num_cpus: o.get_field("numCpus"),
            oauth2: o.get_field("oauth2"),
            placement_labels: o.get_field("placementLabels"),
            publisher_names: o.get_field("publisherNames"),
            publishers: o.get_field("publishers"),
            registrations: o.get_field("registrations"),
            replicas: o.get_field("replicas"),
            tags: o.get_field("tags"),
            template_name: o.get_field("templateName"),
            tenant_url: o.get_field("tenantUrl"),
            wizard_path: o.get_field("wizardPath"),
        }
    }
}
