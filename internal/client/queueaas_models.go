package client

type QueueAASRequest struct {
	BaseRequest
	Name         string `json:"name"`
	NodeNumber   int64  `json:"node_number"`
	NodeSize     string `json:"node_size"`
	SizeGB       int64  `json:"size_gb"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	KafkaVersion string `json:"kafka_version"`
	IsHA         bool   `json:"is_ha"`
	IsPublic     bool   `json:"is_public"`
	NetworkUUID  string `json:"private_network_id"`
}

type QueueAASResponse struct {
	baseResponse
	Name         string `json:"name"`
	NodeNumber   int64  `json:"node_number"`
	NodeSize     string `json:"node_size"`
	SizeGB       int64  `json:"size_gb"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	KafkaVersion string `json:"kafka_version"`
	IsHA         bool   `json:"is_ha"`
	IsPublic     bool   `json:"is_public"`
	NetworkUUID  string `json:"private_network_id"`
}

type QueueAASCredentials struct {
	BootstrapServers string `kv:"BOOTSTRAP_SERVERS"`
	Username         string `kv:"USERNAME"`
	Password         string `kv:"PASSWORD"`
	KafkaCert        string `kv:"KAFKA_CERT"`
	Region           string `kv:"REGION"`
	ServiceUUID      string `kv:"SERVICE_UUID"`
}

func (r QueueAASRequest) canCreateConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/queueaas/cancreate",
		ResourceName: "queueaas",
		Headers:      nil,
	}
}

func (r QueueAASRequest) createConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/queueaas/create",
		ResourceName: "queueaas",
		Headers:      nil,
	}
}
