package client

type LoadBalancerRequest struct {
	BaseRequest
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsPublic bool   `json:"is_public"`
	Port     int64  `json:"port"`
}

type LoadBalancerResponse struct {
	baseResponse
	Name            string `json:"name"`
	Type            string `json:"type"`
	IsPublic        bool   `json:"is_public"`
	Port            int64  `json:"listener_port"`
	Scheme          string `json:"scheme"`
	DNSName         string `json:"dns_name"`
	ARN             string `json:"arn"`
	TargetGroupARN  string `json:"target_group_arn"`
	SecurityGroupID string `json:"security_group_id"`
	VPCID           string `json:"vpc_id"`
	State           string `json:"state"`
}

// type LoadBalancerCredentials struct {
// }

func (r LoadBalancerRequest) canCreateConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/loadbalancer/cancreate",
		ResourceName: "loadbalancer",
		Headers:      nil,
	}
}

func (r LoadBalancerRequest) createConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/loadbalancer/create",
		ResourceName: "loadbalancer",
		Headers:      nil,
	}
}
