package client

type PublicGatewayRequest struct {
	BaseRequest
	Name        string `json:"name"`
	NetworkUUID string `json:"network_uuid"`
}

type PublicGatewayResponse struct {
	baseResponse
	Name           string `json:"name"`
	NetworkUUID    string `json:"network_uuid"`
	VPCID          string `json:"vpc_id"`
	NATGatewayID   string `json:"nat_gateway_id"`
	AllocationID   string `json:"allocation_id"`
	PublicSubnetID string `json:"public_subnet_id"`
	State          string `json:"state"`
	EgressEnabled  bool   `json:"egress_enabled"`
	BytesOutMonth  int64  `json:"bytes_out_month"`
}

// type PublicGatewayCredentials struct {
// }

func (r PublicGatewayRequest) canCreateConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/publicgateway/cancreate",
		ResourceName: "publicgateway",
		Headers:      nil,
	}
}

func (r PublicGatewayRequest) createConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/publicgateway/create",
		ResourceName: "publicgateway",
		Headers:      nil,
	}
}
