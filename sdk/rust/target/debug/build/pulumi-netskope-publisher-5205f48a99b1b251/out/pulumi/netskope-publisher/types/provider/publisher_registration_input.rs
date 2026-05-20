#[derive(pulumi_gestalt_rust::__private::serde::Deserialize, pulumi_gestalt_rust::__private::serde::Serialize, pulumi_gestalt_rust::__private::bon::Builder, Debug, PartialEq, Clone)]
#[builder(finish_fn = build_struct)]
#[allow(dead_code)]
#[allow(clippy::doc_lazy_continuation, clippy::tabs_in_doc_comments, clippy::should_implement_trait)]
pub struct PublisherRegistrationInput {
    #[builder(into)]
    #[serde(rename = "existedBefore")]
    pub r#existed_before: Option<bool>,
    #[builder(into)]
    #[serde(rename = "publisherId")]
    pub r#publisher_id: i32,
    #[builder(into)]
    #[serde(rename = "registrationToken")]
    pub r#registration_token: String,
}
