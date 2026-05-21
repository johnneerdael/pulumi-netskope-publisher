#[allow(
    clippy::doc_lazy_continuation,
    clippy::tabs_in_doc_comments,
    clippy::should_implement_trait
)]
pub mod hyperv_publisher {
    #[derive(pulumi_gestalt_rust::__private::bon::Builder)]
    #[builder(finish_fn = build_struct)]
    #[allow(dead_code)]
    pub struct HypervPublisherArgs {
        #[builder(into, default)]
        pub api_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub auth_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub auto_start_action: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub auto_stop_action: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub bearer_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub dynamic_memory: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub enable_experimental_hyperv: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub generation: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into)]
        pub hard_drives: pulumi_gestalt_rust::Input<
            Vec<super::types::provider::HypervHardDrive>,
        >,
        #[builder(into, default)]
        pub maximum_memory: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub memory_size: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub minimum_memory: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub name_prefix: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub names: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into, default)]
        pub oauth2: pulumi_gestalt_rust::Input<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        #[builder(into, default)]
        pub placement_labels: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into, default)]
        pub processor_count: pulumi_gestalt_rust::Input<Option<i32>>,
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
        #[builder(into)]
        pub switch_name: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub tags: pulumi_gestalt_rust::Input<
            Option<std::collections::HashMap<String, String>>,
        >,
        #[builder(into, default)]
        pub tenant_url: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub wizard_path: pulumi_gestalt_rust::Input<Option<String>>,
    }
    #[allow(dead_code)]
    pub struct HypervPublisherResult {
        /// Pulumi ID is the provider-assigned unique ID for this managed resource.
        /// It is set during deployments and may be missing (unknown) during planning phases.
        pub id: pulumi_gestalt_rust::Output<String>,
        /// Pulumi URN is the stable logical identity of this resource in the Pulumi stack.
        pub urn: pulumi_gestalt_rust::Output<String>,
        pub api_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub auth_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub auto_start_action: pulumi_gestalt_rust::Output<Option<String>>,
        pub auto_stop_action: pulumi_gestalt_rust::Output<Option<String>>,
        pub bearer_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub dynamic_memory: pulumi_gestalt_rust::Output<Option<bool>>,
        pub enable_experimental_hyperv: pulumi_gestalt_rust::Output<Option<bool>>,
        pub generation: pulumi_gestalt_rust::Output<Option<i32>>,
        pub hard_drives: pulumi_gestalt_rust::Output<
            Vec<super::types::provider::HypervHardDrive>,
        >,
        pub maximum_memory: pulumi_gestalt_rust::Output<Option<i32>>,
        pub memory_size: pulumi_gestalt_rust::Output<Option<i32>>,
        pub minimum_memory: pulumi_gestalt_rust::Output<Option<i32>>,
        pub name_prefix: pulumi_gestalt_rust::Output<Option<String>>,
        pub names: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub oauth2: pulumi_gestalt_rust::Output<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        pub placement_labels: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub processor_count: pulumi_gestalt_rust::Output<Option<i32>>,
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
        pub switch_name: pulumi_gestalt_rust::Output<String>,
        pub tags: pulumi_gestalt_rust::Output<
            Option<std::collections::HashMap<String, String>>,
        >,
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
        args: HypervPublisherArgs,
    ) -> HypervPublisherResult {
        __create(ctx, name, args, None)
    }
    ///
    /// Same as `create`, but with additional generic options that control the behavior of the resource registration.
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create_with_options(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: HypervPublisherArgs,
        options: pulumi_gestalt_rust::CustomResourceOptions,
    ) -> HypervPublisherResult {
        __create(ctx, name, args, Some(options))
    }
    #[allow(non_snake_case, unused_imports, dead_code)]
    fn __create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: HypervPublisherArgs,
        options: Option<pulumi_gestalt_rust::CustomResourceOptions>,
    ) -> HypervPublisherResult {
        let api_token_binding = args.api_token.get_output(ctx);
        let auth_mode_binding = args.auth_mode.get_output(ctx);
        let auto_start_action_binding = args.auto_start_action.get_output(ctx);
        let auto_stop_action_binding = args.auto_stop_action.get_output(ctx);
        let bearer_token_binding = args.bearer_token.get_output(ctx);
        let dynamic_memory_binding = args.dynamic_memory.get_output(ctx);
        let enable_experimental_hyperv_binding = args
            .enable_experimental_hyperv
            .get_output(ctx);
        let generation_binding = args.generation.get_output(ctx);
        let hard_drives_binding = args.hard_drives.get_output(ctx);
        let maximum_memory_binding = args.maximum_memory.get_output(ctx);
        let memory_size_binding = args.memory_size.get_output(ctx);
        let minimum_memory_binding = args.minimum_memory.get_output(ctx);
        let name_prefix_binding = args.name_prefix.get_output(ctx);
        let names_binding = args.names.get_output(ctx);
        let oauth2_binding = args.oauth2.get_output(ctx);
        let placement_labels_binding = args.placement_labels.get_output(ctx);
        let processor_count_binding = args.processor_count.get_output(ctx);
        let registrations_binding = args.registrations.get_output(ctx);
        let replicas_binding = args.replicas.get_output(ctx);
        let switch_name_binding = args.switch_name.get_output(ctx);
        let tags_binding = args.tags.get_output(ctx);
        let tenant_url_binding = args.tenant_url.get_output(ctx);
        let wizard_path_binding = args.wizard_path.get_output(ctx);
        let request = pulumi_gestalt_rust::RegisterResourceRequest {
            type_: "netskope-publisher:index:HypervPublisher".into(),
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
                    name: "autoStartAction".into(),
                    value: &auto_start_action_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "autoStopAction".into(),
                    value: &auto_stop_action_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "bearerToken".into(),
                    value: &bearer_token_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "dynamicMemory".into(),
                    value: &dynamic_memory_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "enableExperimentalHyperv".into(),
                    value: &enable_experimental_hyperv_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "generation".into(),
                    value: &generation_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "hardDrives".into(),
                    value: &hard_drives_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "maximumMemory".into(),
                    value: &maximum_memory_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "memorySize".into(),
                    value: &memory_size_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "minimumMemory".into(),
                    value: &minimum_memory_binding.drop_type(),
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
                    name: "oauth2".into(),
                    value: &oauth2_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "placementLabels".into(),
                    value: &placement_labels_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "processorCount".into(),
                    value: &processor_count_binding.drop_type(),
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
                    name: "switchName".into(),
                    value: &switch_name_binding.drop_type(),
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
            ],
            options,
        };
        let o = ctx.register_resource(request);
        HypervPublisherResult {
            id: o.get_id(),
            urn: o.get_urn(),
            api_token: o.get_field("apiToken"),
            auth_mode: o.get_field("authMode"),
            auto_start_action: o.get_field("autoStartAction"),
            auto_stop_action: o.get_field("autoStopAction"),
            bearer_token: o.get_field("bearerToken"),
            dynamic_memory: o.get_field("dynamicMemory"),
            enable_experimental_hyperv: o.get_field("enableExperimentalHyperv"),
            generation: o.get_field("generation"),
            hard_drives: o.get_field("hardDrives"),
            maximum_memory: o.get_field("maximumMemory"),
            memory_size: o.get_field("memorySize"),
            minimum_memory: o.get_field("minimumMemory"),
            name_prefix: o.get_field("namePrefix"),
            names: o.get_field("names"),
            oauth2: o.get_field("oauth2"),
            placement_labels: o.get_field("placementLabels"),
            processor_count: o.get_field("processorCount"),
            publisher_names: o.get_field("publisherNames"),
            publishers: o.get_field("publishers"),
            registrations: o.get_field("registrations"),
            replicas: o.get_field("replicas"),
            switch_name: o.get_field("switchName"),
            tags: o.get_field("tags"),
            tenant_url: o.get_field("tenantUrl"),
            wizard_path: o.get_field("wizardPath"),
        }
    }
}
