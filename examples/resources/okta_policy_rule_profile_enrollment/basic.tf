resource "okta_policy_profile_enrollment" "test" {
  name = "testAcc_replace_with_uuid"
}

resource "okta_policy_rule_profile_enrollment" "test" {
  policy_id           = okta_policy_profile_enrollment.test.id
  unknown_user_action = "REGISTER"
  email_verification  = true
  access              = "ALLOW"
  profile_attributes {
    name     = "email"
    label    = "Email"
    required = true
  }
  enroll_authenticator_types = ["password"]
}


