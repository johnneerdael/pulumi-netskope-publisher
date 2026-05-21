#[allow(
    clippy::doc_lazy_continuation,
    clippy::tabs_in_doc_comments,
    clippy::should_implement_trait
)]
pub mod realtime_protection_policy {
    #[derive(pulumi_gestalt_rust::__private::bon::Builder)]
    #[builder(finish_fn = build_struct)]
    #[allow(dead_code)]
    pub struct RealtimeProtectionPolicyArgs {
        #[builder(into)]
        pub action: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub api_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub app_ids: pulumi_gestalt_rust::Input<Option<Vec<i32>>>,
        #[builder(into, default)]
        pub app_tags: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into, default)]
        pub auth_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub bearer_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into)]
        pub enabled: pulumi_gestalt_rust::Input<bool>,
        #[builder(into, default)]
        pub groups: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into)]
        pub name: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub oauth2: pulumi_gestalt_rust::Input<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        #[builder(into, default)]
        pub policy_group_id: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub policy_group_name: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into)]
        pub tenant_url: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub users: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
    }
    #[allow(dead_code)]
    pub struct RealtimeProtectionPolicyResult {
        /// Pulumi ID is the provider-assigned unique ID for this managed resource.
        /// It is set during deployments and may be missing (unknown) during planning phases.
        pub id: pulumi_gestalt_rust::Output<String>,
        /// Pulumi URN is the stable logical identity of this resource in the Pulumi stack.
        pub urn: pulumi_gestalt_rust::Output<String>,
        pub action: pulumi_gestalt_rust::Output<String>,
        pub api_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub app_ids: pulumi_gestalt_rust::Output<Option<Vec<i32>>>,
        pub app_tags: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub auth_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub bearer_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub enabled: pulumi_gestalt_rust::Output<bool>,
        pub groups: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub name: pulumi_gestalt_rust::Output<String>,
        pub oauth2: pulumi_gestalt_rust::Output<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        pub policy_group_id: pulumi_gestalt_rust::Output<Option<i32>>,
        pub policy_group_name: pulumi_gestalt_rust::Output<Option<String>>,
        pub policy_id: pulumi_gestalt_rust::Output<i32>,
        pub resolved_policy_group_id: pulumi_gestalt_rust::Output<i32>,
        pub tenant_url: pulumi_gestalt_rust::Output<String>,
        pub users: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
    }
    ///
    /// Registers a new resource with the given unique name and arguments
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: RealtimeProtectionPolicyArgs,
    ) -> RealtimeProtectionPolicyResult {
        __create(ctx, name, args, None)
    }
    ///
    /// Same as `create`, but with additional generic options that control the behavior of the resource registration.
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create_with_options(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: RealtimeProtectionPolicyArgs,
        options: pulumi_gestalt_rust::CustomResourceOptions,
    ) -> RealtimeProtectionPolicyResult {
        __create(ctx, name, args, Some(options))
    }
    #[allow(non_snake_case, unused_imports, dead_code)]
    fn __create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: RealtimeProtectionPolicyArgs,
        options: Option<pulumi_gestalt_rust::CustomResourceOptions>,
    ) -> RealtimeProtectionPolicyResult {
        let action_binding = args.action.get_output(ctx);
        let api_token_binding = args.api_token.get_output(ctx);
        let app_ids_binding = args.app_ids.get_output(ctx);
        let app_tags_binding = args.app_tags.get_output(ctx);
        let auth_mode_binding = args.auth_mode.get_output(ctx);
        let bearer_token_binding = args.bearer_token.get_output(ctx);
        let enabled_binding = args.enabled.get_output(ctx);
        let groups_binding = args.groups.get_output(ctx);
        let name_binding = args.name.get_output(ctx);
        let oauth2_binding = args.oauth2.get_output(ctx);
        let policy_group_id_binding = args.policy_group_id.get_output(ctx);
        let policy_group_name_binding = args.policy_group_name.get_output(ctx);
        let tenant_url_binding = args.tenant_url.get_output(ctx);
        let users_binding = args.users.get_output(ctx);
        let request = pulumi_gestalt_rust::RegisterResourceRequest {
            type_: "netskope-publisher:index:RealtimeProtectionPolicy".into(),
            name: name.to_string(),
            version: super::get_version(),
            object: &[
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "action".into(),
                    value: &action_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "apiToken".into(),
                    value: &api_token_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "appIds".into(),
                    value: &app_ids_binding.drop_type(),
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
                    name: "enabled".into(),
                    value: &enabled_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "groups".into(),
                    value: &groups_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "name".into(),
                    value: &name_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "oauth2".into(),
                    value: &oauth2_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "policyGroupId".into(),
                    value: &policy_group_id_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "policyGroupName".into(),
                    value: &policy_group_name_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "tenantUrl".into(),
                    value: &tenant_url_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "users".into(),
                    value: &users_binding.drop_type(),
                },
            ],
            options,
        };
        let o = ctx.register_resource(request);
        RealtimeProtectionPolicyResult {
            id: o.get_id(),
            urn: o.get_urn(),
            action: o.get_field("action"),
            api_token: o.get_field("apiToken"),
            app_ids: o.get_field("appIds"),
            app_tags: o.get_field("appTags"),
            auth_mode: o.get_field("authMode"),
            bearer_token: o.get_field("bearerToken"),
            enabled: o.get_field("enabled"),
            groups: o.get_field("groups"),
            name: o.get_field("name"),
            oauth2: o.get_field("oauth2"),
            policy_group_id: o.get_field("policyGroupId"),
            policy_group_name: o.get_field("policyGroupName"),
            policy_id: o.get_field("policyId"),
            resolved_policy_group_id: o.get_field("resolvedPolicyGroupId"),
            tenant_url: o.get_field("tenantUrl"),
            users: o.get_field("users"),
        }
    }
}
