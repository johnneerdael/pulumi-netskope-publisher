#[allow(
    clippy::doc_lazy_continuation,
    clippy::tabs_in_doc_comments,
    clippy::should_implement_trait
)]
pub mod private_app {
    #[derive(pulumi_gestalt_rust::__private::bon::Builder)]
    #[builder(finish_fn = build_struct)]
    #[allow(dead_code)]
    pub struct PrivateAppArgs {
        #[builder(into, default)]
        pub adopt_existing: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub api_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into)]
        pub app_name: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub app_type: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub auth_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub bearer_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into)]
        pub clientless_access: pulumi_gestalt_rust::Input<bool>,
        #[builder(into)]
        pub host: pulumi_gestalt_rust::Input<String>,
        #[builder(into)]
        pub is_user_portal_app: pulumi_gestalt_rust::Input<bool>,
        #[builder(into, default)]
        pub oauth2: pulumi_gestalt_rust::Input<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        #[builder(into)]
        pub protocols: pulumi_gestalt_rust::Input<
            Vec<super::types::provider::PrivateAppProtocol>,
        >,
        #[builder(into, default)]
        pub publishers: pulumi_gestalt_rust::Input<
            Option<Vec<super::types::provider::PrivateAppPublisher>>,
        >,
        #[builder(into, default)]
        pub tags: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into)]
        pub tenant_url: pulumi_gestalt_rust::Input<String>,
        #[builder(into)]
        pub trust_self_signed_certs: pulumi_gestalt_rust::Input<bool>,
        #[builder(into)]
        pub use_publisher_dns: pulumi_gestalt_rust::Input<bool>,
    }
    #[allow(dead_code)]
    pub struct PrivateAppResult {
        /// Pulumi ID is the provider-assigned unique ID for this managed resource.
        /// It is set during deployments and may be missing (unknown) during planning phases.
        pub id: pulumi_gestalt_rust::Output<String>,
        /// Pulumi URN is the stable logical identity of this resource in the Pulumi stack.
        pub urn: pulumi_gestalt_rust::Output<String>,
        pub adopt_existing: pulumi_gestalt_rust::Output<Option<bool>>,
        pub api_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub app_id: pulumi_gestalt_rust::Output<i32>,
        pub app_name: pulumi_gestalt_rust::Output<String>,
        pub app_type: pulumi_gestalt_rust::Output<Option<String>>,
        pub auth_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub bearer_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub clientless_access: pulumi_gestalt_rust::Output<bool>,
        pub host: pulumi_gestalt_rust::Output<String>,
        pub is_user_portal_app: pulumi_gestalt_rust::Output<bool>,
        pub oauth2: pulumi_gestalt_rust::Output<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        pub protocols: pulumi_gestalt_rust::Output<
            Vec<super::types::provider::PrivateAppProtocol>,
        >,
        pub publishers: pulumi_gestalt_rust::Output<
            Option<Vec<super::types::provider::PrivateAppPublisher>>,
        >,
        pub tags: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub tenant_url: pulumi_gestalt_rust::Output<String>,
        pub trust_self_signed_certs: pulumi_gestalt_rust::Output<bool>,
        pub use_publisher_dns: pulumi_gestalt_rust::Output<bool>,
    }
    ///
    /// Registers a new resource with the given unique name and arguments
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: PrivateAppArgs,
    ) -> PrivateAppResult {
        __create(ctx, name, args, None)
    }
    ///
    /// Same as `create`, but with additional generic options that control the behavior of the resource registration.
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create_with_options(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: PrivateAppArgs,
        options: pulumi_gestalt_rust::CustomResourceOptions,
    ) -> PrivateAppResult {
        __create(ctx, name, args, Some(options))
    }
    #[allow(non_snake_case, unused_imports, dead_code)]
    fn __create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: PrivateAppArgs,
        options: Option<pulumi_gestalt_rust::CustomResourceOptions>,
    ) -> PrivateAppResult {
        let adopt_existing_binding = args.adopt_existing.get_output(ctx);
        let api_token_binding = args.api_token.get_output(ctx);
        let app_name_binding = args.app_name.get_output(ctx);
        let app_type_binding = args.app_type.get_output(ctx);
        let auth_mode_binding = args.auth_mode.get_output(ctx);
        let bearer_token_binding = args.bearer_token.get_output(ctx);
        let clientless_access_binding = args.clientless_access.get_output(ctx);
        let host_binding = args.host.get_output(ctx);
        let is_user_portal_app_binding = args.is_user_portal_app.get_output(ctx);
        let oauth2_binding = args.oauth2.get_output(ctx);
        let protocols_binding = args.protocols.get_output(ctx);
        let publishers_binding = args.publishers.get_output(ctx);
        let tags_binding = args.tags.get_output(ctx);
        let tenant_url_binding = args.tenant_url.get_output(ctx);
        let trust_self_signed_certs_binding = args
            .trust_self_signed_certs
            .get_output(ctx);
        let use_publisher_dns_binding = args.use_publisher_dns.get_output(ctx);
        let request = pulumi_gestalt_rust::RegisterResourceRequest {
            type_: "netskope-publisher:index:PrivateApp".into(),
            name: name.to_string(),
            version: super::get_version(),
            object: &[
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "adoptExisting".into(),
                    value: &adopt_existing_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "apiToken".into(),
                    value: &api_token_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "appName".into(),
                    value: &app_name_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "appType".into(),
                    value: &app_type_binding.drop_type(),
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
                    name: "clientlessAccess".into(),
                    value: &clientless_access_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "host".into(),
                    value: &host_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "isUserPortalApp".into(),
                    value: &is_user_portal_app_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "oauth2".into(),
                    value: &oauth2_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "protocols".into(),
                    value: &protocols_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "publishers".into(),
                    value: &publishers_binding.drop_type(),
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
                    name: "trustSelfSignedCerts".into(),
                    value: &trust_self_signed_certs_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "usePublisherDns".into(),
                    value: &use_publisher_dns_binding.drop_type(),
                },
            ],
            options,
        };
        let o = ctx.register_resource(request);
        PrivateAppResult {
            id: o.get_id(),
            urn: o.get_urn(),
            adopt_existing: o.get_field("adoptExisting"),
            api_token: o.get_field("apiToken"),
            app_id: o.get_field("appId"),
            app_name: o.get_field("appName"),
            app_type: o.get_field("appType"),
            auth_mode: o.get_field("authMode"),
            bearer_token: o.get_field("bearerToken"),
            clientless_access: o.get_field("clientlessAccess"),
            host: o.get_field("host"),
            is_user_portal_app: o.get_field("isUserPortalApp"),
            oauth2: o.get_field("oauth2"),
            protocols: o.get_field("protocols"),
            publishers: o.get_field("publishers"),
            tags: o.get_field("tags"),
            tenant_url: o.get_field("tenantUrl"),
            trust_self_signed_certs: o.get_field("trustSelfSignedCerts"),
            use_publisher_dns: o.get_field("usePublisherDns"),
        }
    }
}
