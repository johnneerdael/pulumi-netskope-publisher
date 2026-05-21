#[allow(
    clippy::doc_lazy_continuation,
    clippy::tabs_in_doc_comments,
    clippy::should_implement_trait
)]
pub mod proxmoxve_publisher {
    #[derive(pulumi_gestalt_rust::__private::bon::Builder)]
    #[builder(finish_fn = build_struct)]
    #[allow(dead_code)]
    pub struct ProxmoxvePublisherArgs {
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
        pub clone_node_name: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub cpu_cores: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into)]
        pub datastore_id: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub delete_default_user: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub disk_size: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub full_clone: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub gateway: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub guest_network_interface: pulumi_gestalt_rust::Input<
            Option<super::types::provider::GuestNetworkInterface>,
        >,
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
        pub ip_address: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub memory: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub name_prefix: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub names: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into, default)]
        pub nameservers: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into, default)]
        pub network_bridge: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub network_model: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into)]
        pub node_name: pulumi_gestalt_rust::Input<String>,
        #[builder(into, default)]
        pub nonat: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub oauth2: pulumi_gestalt_rust::Input<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        #[builder(into, default)]
        pub on_boot: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub placement_labels: pulumi_gestalt_rust::Input<Option<Vec<String>>>,
        #[builder(into, default)]
        pub pool_id: pulumi_gestalt_rust::Input<Option<String>>,
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
        pub started: pulumi_gestalt_rust::Input<Option<bool>>,
        #[builder(into, default)]
        pub tags: pulumi_gestalt_rust::Input<
            Option<std::collections::HashMap<String, String>>,
        >,
        #[builder(into)]
        pub template_vm_id: pulumi_gestalt_rust::Input<i32>,
        #[builder(into, default)]
        pub tenant_url: pulumi_gestalt_rust::Input<Option<String>>,
        #[builder(into, default)]
        pub vlan_id: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub vm_id: pulumi_gestalt_rust::Input<Option<i32>>,
        #[builder(into, default)]
        pub wizard_path: pulumi_gestalt_rust::Input<Option<String>>,
    }
    #[allow(dead_code)]
    pub struct ProxmoxvePublisherResult {
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
        pub clone_node_name: pulumi_gestalt_rust::Output<Option<String>>,
        pub cpu_cores: pulumi_gestalt_rust::Output<Option<i32>>,
        pub datastore_id: pulumi_gestalt_rust::Output<String>,
        pub delete_default_user: pulumi_gestalt_rust::Output<Option<bool>>,
        pub disk_size: pulumi_gestalt_rust::Output<Option<i32>>,
        pub full_clone: pulumi_gestalt_rust::Output<Option<bool>>,
        pub gateway: pulumi_gestalt_rust::Output<Option<String>>,
        pub guest_network_interface: pulumi_gestalt_rust::Output<
            Option<super::types::provider::GuestNetworkInterface>,
        >,
        pub install_user: pulumi_gestalt_rust::Output<Option<String>>,
        pub install_user_password: pulumi_gestalt_rust::Output<Option<String>>,
        pub install_user_password_is_hash: pulumi_gestalt_rust::Output<Option<bool>>,
        pub install_user_ssh_authorized_keys: pulumi_gestalt_rust::Output<
            Option<Vec<String>>,
        >,
        pub ip_address: pulumi_gestalt_rust::Output<Option<String>>,
        pub memory: pulumi_gestalt_rust::Output<Option<i32>>,
        pub name_prefix: pulumi_gestalt_rust::Output<Option<String>>,
        pub names: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub nameservers: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub network_bridge: pulumi_gestalt_rust::Output<Option<String>>,
        pub network_model: pulumi_gestalt_rust::Output<Option<String>>,
        pub node_name: pulumi_gestalt_rust::Output<String>,
        pub nonat: pulumi_gestalt_rust::Output<Option<bool>>,
        pub oauth2: pulumi_gestalt_rust::Output<
            Option<super::types::provider::NetskopeOAuth2Args>,
        >,
        pub on_boot: pulumi_gestalt_rust::Output<Option<bool>>,
        pub placement_labels: pulumi_gestalt_rust::Output<Option<Vec<String>>>,
        pub pool_id: pulumi_gestalt_rust::Output<Option<String>>,
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
        pub started: pulumi_gestalt_rust::Output<Option<bool>>,
        pub tags: pulumi_gestalt_rust::Output<
            Option<std::collections::HashMap<String, String>>,
        >,
        pub template_vm_id: pulumi_gestalt_rust::Output<i32>,
        pub tenant_url: pulumi_gestalt_rust::Output<Option<String>>,
        pub vlan_id: pulumi_gestalt_rust::Output<Option<i32>>,
        pub vm_id: pulumi_gestalt_rust::Output<Option<i32>>,
        pub wizard_path: pulumi_gestalt_rust::Output<Option<String>>,
    }
    ///
    /// Registers a new resource with the given unique name and arguments
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: ProxmoxvePublisherArgs,
    ) -> ProxmoxvePublisherResult {
        __create(ctx, name, args, None)
    }
    ///
    /// Same as `create`, but with additional generic options that control the behavior of the resource registration.
    ///
    #[allow(non_snake_case, dead_code)]
    pub fn create_with_options(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: ProxmoxvePublisherArgs,
        options: pulumi_gestalt_rust::CustomResourceOptions,
    ) -> ProxmoxvePublisherResult {
        __create(ctx, name, args, Some(options))
    }
    #[allow(non_snake_case, unused_imports, dead_code)]
    fn __create(
        ctx: &pulumi_gestalt_rust::Context,
        name: &str,
        args: ProxmoxvePublisherArgs,
        options: Option<pulumi_gestalt_rust::CustomResourceOptions>,
    ) -> ProxmoxvePublisherResult {
        let api_token_binding = args.api_token.get_output(ctx);
        let auth_mode_binding = args.auth_mode.get_output(ctx);
        let bearer_token_binding = args.bearer_token.get_output(ctx);
        let bootstrap_binding = args.bootstrap.get_output(ctx);
        let bootstrap_url_binding = args.bootstrap_url.get_output(ctx);
        let clone_node_name_binding = args.clone_node_name.get_output(ctx);
        let cpu_cores_binding = args.cpu_cores.get_output(ctx);
        let datastore_id_binding = args.datastore_id.get_output(ctx);
        let delete_default_user_binding = args.delete_default_user.get_output(ctx);
        let disk_size_binding = args.disk_size.get_output(ctx);
        let full_clone_binding = args.full_clone.get_output(ctx);
        let gateway_binding = args.gateway.get_output(ctx);
        let guest_network_interface_binding = args
            .guest_network_interface
            .get_output(ctx);
        let install_user_binding = args.install_user.get_output(ctx);
        let install_user_password_binding = args.install_user_password.get_output(ctx);
        let install_user_password_is_hash_binding = args
            .install_user_password_is_hash
            .get_output(ctx);
        let install_user_ssh_authorized_keys_binding = args
            .install_user_ssh_authorized_keys
            .get_output(ctx);
        let ip_address_binding = args.ip_address.get_output(ctx);
        let memory_binding = args.memory.get_output(ctx);
        let name_prefix_binding = args.name_prefix.get_output(ctx);
        let names_binding = args.names.get_output(ctx);
        let nameservers_binding = args.nameservers.get_output(ctx);
        let network_bridge_binding = args.network_bridge.get_output(ctx);
        let network_model_binding = args.network_model.get_output(ctx);
        let node_name_binding = args.node_name.get_output(ctx);
        let nonat_binding = args.nonat.get_output(ctx);
        let oauth2_binding = args.oauth2.get_output(ctx);
        let on_boot_binding = args.on_boot.get_output(ctx);
        let placement_labels_binding = args.placement_labels.get_output(ctx);
        let pool_id_binding = args.pool_id.get_output(ctx);
        let registrations_binding = args.registrations.get_output(ctx);
        let replicas_binding = args.replicas.get_output(ctx);
        let started_binding = args.started.get_output(ctx);
        let tags_binding = args.tags.get_output(ctx);
        let template_vm_id_binding = args.template_vm_id.get_output(ctx);
        let tenant_url_binding = args.tenant_url.get_output(ctx);
        let vlan_id_binding = args.vlan_id.get_output(ctx);
        let vm_id_binding = args.vm_id.get_output(ctx);
        let wizard_path_binding = args.wizard_path.get_output(ctx);
        let request = pulumi_gestalt_rust::RegisterResourceRequest {
            type_: "netskope-publisher:index:ProxmoxvePublisher".into(),
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
                    name: "cloneNodeName".into(),
                    value: &clone_node_name_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "cpuCores".into(),
                    value: &cpu_cores_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "datastoreId".into(),
                    value: &datastore_id_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "deleteDefaultUser".into(),
                    value: &delete_default_user_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "diskSize".into(),
                    value: &disk_size_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "fullClone".into(),
                    value: &full_clone_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "gateway".into(),
                    value: &gateway_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "guestNetworkInterface".into(),
                    value: &guest_network_interface_binding.drop_type(),
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
                    name: "ipAddress".into(),
                    value: &ip_address_binding.drop_type(),
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
                    name: "nameservers".into(),
                    value: &nameservers_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "networkBridge".into(),
                    value: &network_bridge_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "networkModel".into(),
                    value: &network_model_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "nodeName".into(),
                    value: &node_name_binding.drop_type(),
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
                    name: "onBoot".into(),
                    value: &on_boot_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "placementLabels".into(),
                    value: &placement_labels_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "poolId".into(),
                    value: &pool_id_binding.drop_type(),
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
                    name: "started".into(),
                    value: &started_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "tags".into(),
                    value: &tags_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "templateVmId".into(),
                    value: &template_vm_id_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "tenantUrl".into(),
                    value: &tenant_url_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "vlanId".into(),
                    value: &vlan_id_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "vmId".into(),
                    value: &vm_id_binding.drop_type(),
                },
                pulumi_gestalt_rust::ResourceRequestObjectField {
                    name: "wizardPath".into(),
                    value: &wizard_path_binding.drop_type(),
                },
            ],
            options,
        };
        let o = ctx.register_resource(request);
        ProxmoxvePublisherResult {
            id: o.get_id(),
            urn: o.get_urn(),
            api_token: o.get_field("apiToken"),
            auth_mode: o.get_field("authMode"),
            bearer_token: o.get_field("bearerToken"),
            bootstrap: o.get_field("bootstrap"),
            bootstrap_url: o.get_field("bootstrapUrl"),
            clone_node_name: o.get_field("cloneNodeName"),
            cpu_cores: o.get_field("cpuCores"),
            datastore_id: o.get_field("datastoreId"),
            delete_default_user: o.get_field("deleteDefaultUser"),
            disk_size: o.get_field("diskSize"),
            full_clone: o.get_field("fullClone"),
            gateway: o.get_field("gateway"),
            guest_network_interface: o.get_field("guestNetworkInterface"),
            install_user: o.get_field("installUser"),
            install_user_password: o.get_field("installUserPassword"),
            install_user_password_is_hash: o.get_field("installUserPasswordIsHash"),
            install_user_ssh_authorized_keys: o
                .get_field("installUserSshAuthorizedKeys"),
            ip_address: o.get_field("ipAddress"),
            memory: o.get_field("memory"),
            name_prefix: o.get_field("namePrefix"),
            names: o.get_field("names"),
            nameservers: o.get_field("nameservers"),
            network_bridge: o.get_field("networkBridge"),
            network_model: o.get_field("networkModel"),
            node_name: o.get_field("nodeName"),
            nonat: o.get_field("nonat"),
            oauth2: o.get_field("oauth2"),
            on_boot: o.get_field("onBoot"),
            placement_labels: o.get_field("placementLabels"),
            pool_id: o.get_field("poolId"),
            publisher_names: o.get_field("publisherNames"),
            publishers: o.get_field("publishers"),
            registrations: o.get_field("registrations"),
            replicas: o.get_field("replicas"),
            started: o.get_field("started"),
            tags: o.get_field("tags"),
            template_vm_id: o.get_field("templateVmId"),
            tenant_url: o.get_field("tenantUrl"),
            vlan_id: o.get_field("vlanId"),
            vm_id: o.get_field("vmId"),
            wizard_path: o.get_field("wizardPath"),
        }
    }
}
