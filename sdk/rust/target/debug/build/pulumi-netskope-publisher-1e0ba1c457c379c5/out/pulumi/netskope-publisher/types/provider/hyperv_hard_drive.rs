#[derive(pulumi_gestalt_rust::__private::serde::Deserialize, pulumi_gestalt_rust::__private::serde::Serialize, pulumi_gestalt_rust::__private::bon::Builder, Debug, PartialEq, Clone)]
#[builder(finish_fn = build_struct)]
#[allow(dead_code)]
#[allow(clippy::doc_lazy_continuation, clippy::tabs_in_doc_comments, clippy::should_implement_trait)]
pub struct HypervHardDrive {
    #[builder(into)]
    #[serde(rename = "controllerLocation")]
    pub r#controller_location: Option<i32>,
    #[builder(into)]
    #[serde(rename = "controllerNumber")]
    pub r#controller_number: Option<i32>,
    #[builder(into)]
    #[serde(rename = "controllerType")]
    pub r#controller_type: Option<String>,
    #[builder(into)]
    #[serde(rename = "path")]
    pub r#path: String,
}
