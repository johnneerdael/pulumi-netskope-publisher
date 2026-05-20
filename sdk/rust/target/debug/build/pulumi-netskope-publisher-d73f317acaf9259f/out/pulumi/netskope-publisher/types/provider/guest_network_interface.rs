#[derive(pulumi_gestalt_rust::__private::serde::Deserialize, pulumi_gestalt_rust::__private::serde::Serialize, pulumi_gestalt_rust::__private::bon::Builder, Debug, PartialEq, Clone)]
#[builder(finish_fn = build_struct)]
#[allow(dead_code)]
#[allow(clippy::doc_lazy_continuation, clippy::tabs_in_doc_comments, clippy::should_implement_trait)]
pub struct GuestNetworkInterface {
    #[builder(into)]
    #[serde(rename = "addresses")]
    pub r#addresses: Option<Vec<String>>,
    #[builder(into)]
    #[serde(rename = "dhcp4")]
    pub r#dhcp_4: Option<bool>,
    #[builder(into)]
    #[serde(rename = "gateway4")]
    pub r#gateway_4: Option<String>,
    #[builder(into)]
    #[serde(rename = "mtu")]
    pub r#mtu: Option<i32>,
    #[builder(into)]
    #[serde(rename = "name")]
    pub r#name: String,
    #[builder(into)]
    #[serde(rename = "nameservers")]
    pub r#nameservers: Option<Vec<String>>,
}
