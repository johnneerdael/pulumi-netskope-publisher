#[allow(
    clippy::doc_lazy_continuation,
    clippy::tabs_in_doc_comments,
    clippy::should_implement_trait
)]
pub mod ovh_publisher {
    #[derive(pulumi_gestalt_rust::__private::bon::Builder)]
    #[builder(finish_fn = build_struct)]
    #[allow(dead_code)]
    pub struct OvhPublisherArgs {
        #[builder(into, default)]
        pub api_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub auth_mode: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub bearer_token: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub bootstrap: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub bootstrap_url: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub delete_default_user: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into)]
        pub flavor_id: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub guest_network_interface: pulumi_gestalt_rust::Input<
            Option<super::types::provider::GuestNetworkInterface>,
        >,
        #[builder(into)]
        pub image_id: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub install_user: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub install_user_password: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub install_user_password_is_hash: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub install_user_ssh_authorized_keys: pulumi_gestalt_rust::Input<
            Option<Vec<String>>,
        >,
        #[builder(into, default)]
        pub name_prefix: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub names: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into, default)]
        pub network_id: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub nonat: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub oauth2: pulumi_gestalt_rust::Input<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        #[builder(into, default)]
        pub placement_labels: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into)]
        pub region: pulumi_gestalt_rust::Input<String>,
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
        pub service_name: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub ssh_key_name: pulumi_gestalt_rust::Input<Option<String>>,
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
    pub struct OvhPublisherResult {
        /// Pulumi ID is the provider-assigned unique ID for this managed resource.
        /// It is set during deployments and may be missing (unknown) during planning phases.
        pub id: pulumi_gestalt_rust::Output<String>,
        /// Pulumi URN is the stable logical identity of this resource in the Pulumi stack.
        pub urn: pulumi_gestalt_rust::Output<String>,
        pub api_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub auth_mode: pulumi_gestalt_rust::Output<Option<String>>,
        pub bearer_token: pulumi_gestalt_rust::Output<Option<String>>,
        pub bootstrap: pulumi_gestalt_rust::Output<Option<bool>>,
        pub bootstrap_url: pulumi_gestalt_rust::Output<Option<String>>,
        pub delete_default_user: pulumi_gestalt_rust::Output<Option<bool>>,
        pub flavor_id: pulumi_gestalt_rust::Output<String>,
        pub guest_network_interface: pulumi_gestalt_rust::Output<
            Option<super::types::provider::GuestNetworkInterface>,
        >,
        pub image_id: pulumi_gestalt_rust::Output<String>,
        pub install_user: pulumi_gestalt_rust::Output<Option<String>>,
        pub install_user_password: pulumi_gestalt_rust::Output<Option<String>>,
        pub install_user_password_is_hash: pulumi_gestalt_rust::Output<Option<bool>>,
        pub install_user_ssh_authorized_keys: pulumi_gestalt_rust::Output<
            Option<Vec<String>>,
        >,
        pub name_prefix: pulumi_gestalt_rust::Output<Option<String>>,
        pub names: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub network_id: pulumi_gestalt_rust::Output<Option<String>>,
        pub nonat: pulumi_gestalt_rust::Output<Option<bool>>,
        pub oauth2: pulumi_gestalt_rust::Output<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        pub placement_labels: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub publisher_names: pulumi_gestalt_rust::Output<Vec<String>>,
        pub publishers: pulumi_gestalt_rust::Output<
            std::collections::HashMap<String, String>,
        >,
        pub region: pulumi_gestalt_rust::Output<String>,
        pub registrations: pulumi_gestalt_rust::Output<
            Option<
                std::collections::HashMap<
                    String,
                    super::types::provider::PublisherRegistrationInput,
                >,
            >,
        >,
        pub replicas: pulumi_gestalt_rust::Output<Option<i32>>,
        pub service_name: pulumi_gestalt_rust::Output<String>,
        pub ssh_key_name: pulumi_gestalt_rust::Output<Option<String>>,
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
        args: OvhPublisherArgs,
    ) -> OvhPublisherResult {
        __create(ctx, name, args, None)
    }
    ///
    /// Same as `create`, but with additional generic options that control the behavior of the resource registration.
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create_with_options(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: OvhPublisherArgs,
        options: pulumi_gestalt_rust::CustomResourceOptions,
    ) -> OvhPublisherResult {
        __create(ctx, name, args, Some(options))
    }
    #[allow(non_snake_case, unused_imports, dead_code)]
    fn __create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: OvhPublisherArgs,
        options: Option<pulumi_gestalt_rust::CustomResourceOptions>,
    ) -> OvhPublisherResult {
        let api_token_binding = args.api_token.get_output(ctx);
        let auth_mode_binding = args.auth_mode.get_output(ctx);
        let bearer_token_binding = args.bearer_token.get_output(ctx);
        let bootstrap_binding = args.bootstrap.get_output(ctx);
        let bootstrap_url_binding = args.bootstrap_url.get_output(ctx);
        let delete_default_user_binding = args.delete_default_user.get_output(ctx);
        let flavor_id_binding = args.flavor_id.get_output(ctx);
        let guest_network_interface_binding = args
            .guest_network_interface
            .get_output(ctx);
        let image_id_binding = args.image_id.get_output(ctx);
        let install_user_binding = args.install_user.get_output(ctx);
        let install_user_password_binding = args.install_user_password.get_output(ctx);
        let install_user_password_is_hash_binding = args
            .install_user_password_is_hash
            .get_output(ctx);
        let install_user_ssh_authorized_keys_binding = args
            .install_user_ssh_authorized_keys
            .get_output(ctx);
        let name_prefix_binding = args.name_prefix.get_output(ctx);
        let names_binding = args.names.get_output(ctx);
        let network_id_binding = args.network_id.get_output(ctx);
        let nonat_binding = args.nonat.get_output(ctx);
        let oauth2_binding = args.oauth2.get_output(ctx);
        let placement_labels_binding = args.placement_labels.get_output(ctx);
        let region_binding = args.region.get_output(ctx);
        let registrations_binding = args.registrations.get_output(ctx);
        let replicas_binding = args.replicas.get_output(ctx);
        let service_name_binding = args.service_name.get_output(ctx);
        let ssh_key_name_binding = args.ssh_key_name.get_output(ctx);
        let tags_binding = args.tags.get_output(ctx);
        let tenant_url_binding = args.tenant_url.get_output(ctx);
        let wizard_path_binding = args.wizard_path.get_output(ctx);
        let request = pulumi_gestalt_rust::RegisterResourceRequest {
            type_: "netskope-publisher:index:OvhPublisher".into(),
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
                    name: "bootstrap".into(),
                    value: &bootstrap_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "bootstrapUrl".into(),
                    value: &bootstrap_url_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "deleteDefaultUser".into(),
                    value: &delete_default_user_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "flavorId".into(),
                    value: &flavor_id_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "guestNetworkInterface".into(),
                    value: &guest_network_interface_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "imageId".into(),
                    value: &image_id_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "installUser".into(),
                    value: &install_user_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "installUserPassword".into(),
                    value: &install_user_password_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "installUserPasswordIsHash".into(),
                    value: &install_user_password_is_hash_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "installUserSshAuthorizedKeys".into(),
                    value: &install_user_ssh_authorized_keys_binding.drop_type(),
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
                    name: "networkId".into(),
                    value: &network_id_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "nonat".into(),
                    value: &nonat_binding.drop_type(),
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
                    name: "region".into(),
                    value: &region_binding.drop_type(),
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
                    name: "serviceName".into(),
                    value: &service_name_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "sshKeyName".into(),
                    value: &ssh_key_name_binding.drop_type(),
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
        OvhPublisherResult {
            id: o.get_id(),
            urn: o.get_urn(),
            api_token: o.get_field("apiToken"),
            auth_mode: o.get_field("authMode"),
            bearer_token: o.get_field("bearerToken"),
            bootstrap: o.get_field("bootstrap"),
            bootstrap_url: o.get_field("bootstrapUrl"),
            delete_default_user: o.get_field("deleteDefaultUser"),
            flavor_id: o.get_field("flavorId"),
            guest_network_interface: o.get_field("guestNetworkInterface"),
            image_id: o.get_field("imageId"),
            install_user: o.get_field("installUser"),
            install_user_password: o.get_field("installUserPassword"),
            install_user_password_is_hash: o.get_field("installUserPasswordIsHash"),
            install_user_ssh_authorized_keys: o
                .get_field("installUserSshAuthorizedKeys"),
            name_prefix: o.get_field("namePrefix"),
            names: o.get_field("names"),
            network_id: o.get_field("networkId"),
            nonat: o.get_field("nonat"),
            oauth2: o.get_field("oauth2"),
            placement_labels: o.get_field("placementLabels"),
            publisher_names: o.get_field("publisherNames"),
            publishers: o.get_field("publishers"),
            region: o.get_field("region"),
            registrations: o.get_field("registrations"),
            replicas: o.get_field("replicas"),
            service_name: o.get_field("serviceName"),
            ssh_key_name: o.get_field("sshKeyName"),
            tags: o.get_field("tags"),
            tenant_url: o.get_field("tenantUrl"),
            wizard_path: o.get_field("wizardPath"),
        }
    }
}
