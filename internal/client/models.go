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
