package client

type PrivateNetworkRequest struct {
	BaseRequest
	Name           string `json:"name"`
	CIDR           string `json:"cidr"`
	AutoAssignCIDR bool   `json:"auto_assign_cidr"`
}

type PrivateNetworkResponse struct {
	baseResponse
	Name           string `json:"name"`
	AutoAssignCIDR bool   `json:"auto_assign_cidr"`
	CIDR           string `json:"cidr"`
	VPCID          string `json:"vpc_id"`
	Subnets        []struct {
		AZ   string `json:"az"`
		ID   string `json:"subnet_id"`
		CIDR string `json:"cidr"`
	} `json:"subnets"`
	State string `json:"state"`
}

type PrivateNetworkCredentials struct {
	VPCID         string   `kv:"NETWORK_ID"`
	CIDR          string   `kv:"NETWORK_CIDR"`
	SubnetIDs     []string `kv:"NETWORK_SUBNET_IDS,csv"`
	NetworkRegion string   `kv:"NETWORK_REGION"`
	ServiceUUID   string   `kv:"SERVICE_UUID"`
}
