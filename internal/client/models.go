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
