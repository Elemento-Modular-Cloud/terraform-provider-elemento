package client

type PublicIPRequest struct {
	BaseRequest
	Name string `json:"name"`
}

type PublicIPResponse struct {
	baseResponse
	Name                         string `json:"name"`
	PublicIP                     string `json:"public_ip"`
	AllocationID                 string `json:"allocation_id"`
	Domain                       string `json:"domain"`
	AssociationID                string `json:"association_id"`
	AssociatedInstanceID         string `json:"associated_instance_id"`
	AssociatedNetworkInterfaceID string `json:"associated_network_interface_id"`
	State                        string `json:"state"`
}

// type PublicIPCredentials struct {
// }

func (r PublicIPRequest) canCreateConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/publicip/cancreate",
		ResourceName: "publicip",
		Headers:      nil,
	}
}

func (r PublicIPRequest) createConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/publicip/create",
		ResourceName: "publicip",
		Headers:      nil,
	}
}
