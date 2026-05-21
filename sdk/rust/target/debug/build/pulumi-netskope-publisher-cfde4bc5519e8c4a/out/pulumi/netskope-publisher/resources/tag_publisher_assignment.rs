#[allow(
    clippy::doc_lazy_continuation,
    clippy::tabs_in_doc_comments,
    clippy::should_implement_trait
)]
pub mod tag_publisher_assignment {
    #[derive(pulumi_gestalt_rust::__private::bon::Builder)]
    #[builder(finish_fn = build_struct)]
    #[allow(dead_code)]
    pub struct TagPublisherAssignmentArgs {
        #[builder(into, default)]
        pub api_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into)]
        pub app_tags: pulumi_gestalt_rust::Input<Vec<String>>,
        #[builder(into, default)]
        pub auth_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub bearer_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub match_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub oauth2: pulumi_gestalt_rust::Input<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        #[builder(into)]
        pub publisher_placement_labels: pulumi_gestalt_rust::Input<Vec<String>>,
        #[builder(into)]
        pub publishers: pulumi_gestalt_rust::Input<
            std::collections::HashMap<
                String,
                super::types::provider::PublisherAssignmentInput,
            >,
        >,
        #[builder(into)]
        pub tenant_url: pulumi_gestalt_rust::Input<String>,
    }
    #[allow(dead_code)]
    pub struct TagPublisherAssignmentResult {
        /// Pulumi ID is the provider-assigned unique ID for this managed resource.
        /// It is set during deployments and may be missing (unknown) during planning phases.
        pub id: pulumi_gestalt_rust::Output<String>,
        /// Pulumi URN is the stable logical identity of this resource in the Pulumi stack.
        pub urn: pulumi_gestalt_rust::Output<String>,
        pub api_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub app_tags: pulumi_gestalt_rust::Output<Vec<String>>,
        pub auth_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub bearer_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub match_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub matched_apps: pulumi_gestalt_rust::Output<Vec<String>>,
        pub oauth2: pulumi_gestalt_rust::Output<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        pub publisher_placement_labels: pulumi_gestalt_rust::Output<Vec<String>>,
        pub publishers: pulumi_gestalt_rust::Output<
            std::collections::HashMap<
                String,
                super::types::provider::PublisherAssignmentInput,
            >,
        >,
        pub selected_publishers: pulumi_gestalt_rust::Output<Vec<i32>>,
        pub tenant_url: pulumi_gestalt_rust::Output<String>,
    }
    ///
    /// Registers a new resource with the given unique name and arguments
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: TagPublisherAssignmentArgs,
    ) -> TagPublisherAssignmentResult {
        __create(ctx, name, args, None)
    }
    ///
    /// Same as `create`, but with additional generic options that control the behavior of the resource registration.
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create_with_options(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: TagPublisherAssignmentArgs,
        options: pulumi_gestalt_rust::CustomResourceOptions,
    ) -> TagPublisherAssignmentResult {
        __create(ctx, name, args, Some(options))
    }
    #[allow(non_snake_case, unused_imports, dead_code)]
    fn __create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: TagPublisherAssignmentArgs,
        options: Option<pulumi_gestalt_rust::CustomResourceOptions>,
    ) -> TagPublisherAssignmentResult {
        let api_token_binding = args.api_token.get_output(ctx);
        let app_tags_binding = args.app_tags.get_output(ctx);
        let auth_mode_binding = args.auth_mode.get_output(ctx);
        let bearer_token_binding = args.bearer_token.get_output(ctx);
        let match_mode_binding = args.match_mode.get_output(ctx);
        let oauth2_binding = args.oauth2.get_output(ctx);
        let publisher_placement_labels_binding = args
            .publisher_placement_labels
            .get_output(ctx);
        let publishers_binding = args.publishers.get_output(ctx);
        let tenant_url_binding = args.tenant_url.get_output(ctx);
        let request = pulumi_gestalt_rust::RegisterResourceRequest {
            type_: "netskope-publisher:index:TagPublisherAssignment".into(),
            name: name.to_string(),
            version: super::get_version(),
            object: &[
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "apiToken".into(),
                    value: &api_token_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "appTags".into(),
                    value: &app_tags_binding.drop_type(),
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
                    name: "matchMode".into(),
                    value: &match_mode_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "oauth2".into(),
                    value: &oauth2_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "publisherPlacementLabels".into(),
                    value: &publisher_placement_labels_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "publishers".into(),
                    value: &publishers_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "tenantUrl".into(),
                    value: &tenant_url_binding.drop_type(),
                },
            ],
            options,
        };
        let o = ctx.register_resource(request);
        TagPublisherAssignmentResult {
            id: o.get_id(),
            urn: o.get_urn(),
            api_token: o.get_field("apiToken"),
            app_tags: o.get_field("appTags"),
            auth_mode: o.get_field("authMode"),
            bearer_token: o.get_field("bearerToken"),
            match_mode: o.get_field("matchMode"),
            matched_apps: o.get_field("matchedApps"),
            oauth2: o.get_field("oauth2"),
            publisher_placement_labels: o.get_field("publisherPlacementLabels"),
            publishers: o.get_field("publishers"),
            selected_publishers: o.get_field("selectedPublishers"),
            tenant_url: o.get_field("tenantUrl"),
        }
    }
}
