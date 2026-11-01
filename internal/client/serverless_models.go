package client

type ServerlessRequest struct {
	BaseRequest
	Name        string            `json:"name"`
	Runtime     string            `json:"runtime"`
	Handler     string            `json:"handler"`
	Memory      int64             `json:"memory"`
	Timeout     int64             `json:"timeout"`
	IsPublic    bool              `json:"is_public"`
	Environment map[string]string `json:"environment"`
	CodeZipB64  string            `json:"code_zip_b64"`
}

type ServerlessResponse struct {
	baseResponse
	Name     string `json:"name"`
	Runtime  string `json:"runtime"`
	Handler  string `json:"handler"`
	Memory   int64  `json:"memory"`
	Timeout  int64  `json:"timeout"`
	IsPublic bool   `json:"is_public"`
	URL      string `json:"url"`
	ARN      string `json:"arn"`
	State    string `json:"state"`
}

type ServerlessCredentials struct {
	AccessKeyID     string `kv:"[A-Z]{3}_ACCESS_KEY_ID,regex"`
	SecretAccessKey string `kv:"[A-Z]{3}_SECRET_ACCESS_KEY,regex"`
	FunctionURL     string `kv:"FUNCTION_URL"`
	FunctionARN     string `kv:"FUNCTION_ARN"`
	AuthType        string `kv:"AUTH_TYPE"`
	Region          string `kv:"REGION"`
	ServiceUUID     string `kv:"SERVICE_UUID"`
}

func (r ServerlessRequest) canCreateConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/serverless/cancreate",
		ResourceName: "serverless",
		Headers:      nil,
	}
}

func (r ServerlessRequest) createConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/serverless/create",
		ResourceName: "serverless",
		Headers:      nil,
	}
}
