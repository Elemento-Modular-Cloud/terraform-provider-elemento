package client

type CreateResponse interface {
    isCreateResponse()
}

type PublicCreateResponse struct {
    PaymentURL  string `json:"payment_url"`
    BillingUUID string `json:"billing_uuid"`
}

type PrivateCreateResponse struct {
    ServiceUUID string `json:"service_uuid"`
    Status      string `json:"status"`
}

func (r PublicCreateResponse) isCreateResponse() {}
func (r PrivateCreateResponse) isCreateResponse() {}

type TargetGrant struct {
	UUID         string `json:"target_id"`
	OrgUUID      string `json:"org_id"`
	TargetType   string `json:"target_type"`
	TargetName   string `json:"target_name"`
	TargetConfig struct {
		IPList   []string `json:"ips"`
		Provider *string  `json:"provider"`
		Host     *string  `json:"host"`
	} `json:"target_config"`
}

type TargetGrantsResponse struct {
	Targets []TargetGrant `json:"targets"`
}

type ActiveTargetsResponse struct {
	Targets []TargetGrant `json:"active_target_grants"`
}

type CanCreateBilling struct {
    BillingUUID string `json:"billing_uuid"`
    PaymentURL  string `json:"payment_url"`
}

type CanCreateResponse struct {
    CanCreate bool               `json:"cancreate"`
    Billing   []CanCreateBilling `json:"billing"`
    Provider  string             `json:"provider"`
    Reasons   map[string]string  `json:"reasons"`
}

type authRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type authResponse struct {
    Username       string `json:"username"`
    AccessToken    string `json:"access_token"`
    RefreshToken   string `json:"refresh_token"`
    AccessTokenIAT int64  `json:"access_token_iat"`
    AccessTokenExp int64  `json:"access_token_exp"`
}

type BaseRequest struct {
	Region string `json:"region"`
	BillingFrequency *string `json:"billing_frequency,omitempty"`
}

type baseResponse struct {
	CreatorUUID  string `json:"client_uuid"`
	BillingUUID  string `json:"billing_uuid"`
	ServiceUUID  string `json:"service_uuid"`
	Region       string `json:"region"`
	CreationDate string `json:"creation_date"`
}
