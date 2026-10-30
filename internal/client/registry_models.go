package client

type RegistryRequest struct {
	BaseRequest
	Name     string `json:"name"`
	SizeGB   int64  `json:"size_gb"`
	IsPublic bool   `json:"is_public"`
}

type RegistryResponse struct {
	baseResponse
	Name          string `json:"name"`
	SizeGB        int64  `json:"size_gb"`
	IsPublic      bool   `json:"is_public"`
	RepositoryURI string `json:"repository_uri"`
	RegistryID    string `json:"registry_id"`
	ImageCount    int64  `json:"image_count"`
	UsedBytes     int64  `json:"used_bytes"`
	State         string `json:"state"`
}

type RegistryCredentials struct {
	AccessKeyID     string `kv:"[A-Z]{3}_ACCESS_KEY_ID,regex"`
	SecretAccessKey string `kv:"[A-Z]{3}_SECRET_ACCESS_KEY,regex"`
	Region          string `kv:"[A-Z]{3}_REGION,regex"`
	RegistryType    string `kv:"REGISTRY_TYPE"`
	RegistryId      string `kv:"REGISTRY_ID"`
	RepositoryURI   string `kv:"REPOSITORY_URI"`
	LoginCommand    string `kv:"LOGIN_COMMAND"`
	ServiceUUID     string `kv:"SERVICE_UUID"`
}

func (r RegistryRequest) canCreateConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/registry/cancreate",
		ResourceName: "registry",
		Headers:      nil,
	}
}

func (r RegistryRequest) createConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/registry/create",
		ResourceName: "registry",
		Headers:      nil,
	}
}
