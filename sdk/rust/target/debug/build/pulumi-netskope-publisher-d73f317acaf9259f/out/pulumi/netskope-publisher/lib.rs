include!("resources/alicloud_publisher.rs");
include!("resources/aws_publisher.rs");
include!("resources/azure_publisher.rs");
include!("resources/esxi_publisher.rs");
include!("resources/gcp_publisher.rs");
include!("resources/hcloud_publisher.rs");
include!("resources/hyperv_publisher.rs");
include!("resources/kubernetes_publisher.rs");
include!("resources/netskope_registration.rs");
include!("resources/nutanix_publisher.rs");
include!("resources/oci_publisher.rs");
include!("resources/openstack_publisher.rs");
include!("resources/ovh_publisher.rs");
include!("resources/scaleway_publisher.rs");
include!("resources/vsphere_publisher.rs");
pub mod provider {
    include!("provider/provider.rs");
}
pub mod functions {}
pub mod types {
    pub mod provider {
        include!("types/provider/azure_marketplace_image.rs");
        include!("types/provider/azure_os_disk.rs");
        include!("types/provider/gcp_service_account.rs");
        include!("types/provider/guest_network_interface.rs");
        include!("types/provider/hyperv_hard_drive.rs");
        include!("types/provider/metadata_options.rs");
        include!("types/provider/netskope_o_auth_2_args.rs");
        include!("types/provider/publisher_registration_input.rs");
        include!("types/provider/registration_record.rs");
    }
}
#[doc(hidden)]
pub mod constants {}
#[unsafe(link_section = "pulumi_gestalt_provider::netskope-publisher")]
#[unsafe(no_mangle)]
#[cfg(target_arch = "wasm32")]
static PULUMI_WASM_PROVIDER_NETSKOPE_PUBLISHER: [u8; 105] = *b"{\"version\":\"0.1.11\",\"pluginDownloadURL\":\"github://api.github.com/johnneerdael/pulumi-netskope-publisher\"}";
pub(crate) fn get_version() -> String {
    "0.1.11".to_string()
}
