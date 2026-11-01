package client

type AIServerlessRequest struct {
	BaseRequest
	Name            string `json:"name"`
	ModelName       string `json:"model_name"`
	EnableStreaming bool   `json:"enable_streaming"`
}

type AIServerlessResponse struct {
	baseResponse
	Name            string `json:"name"`
	ModelName       string `json:"model_name"`
	EnableStreaming bool   `json:"enable_streaming"`
	ModelID         string `json:"model_id"`
	ARN             string `json:"arn"`
	Status          string `json:"status"`
}

type AIServerlessCredentials struct {
	AccessKeyID      string `kv:"[A-Z]{3}_ACCESS_KEY_ID,regex"`
	SecretAccessKey  string `kv:"[A-Z]{3}_SECRET_ACCESS_KEY,regex"`
	Region           string `kv:"BEDROCK_REGION"`
	ModelName        string `kv:"MODEL_NAME"`
	ModelID          string `kv:"MODEL_ID"`
	StreamingEnabled bool   `kv:"STREAMING_ENABLED"`
	ServiceUUID      string `kv:"SERVICE_UUID"`
}

func (r AIServerlessRequest) canCreateConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/aiserverless/cancreate",
		ResourceName: "aiserverless",
		Headers:      nil,
	}
}

func (r AIServerlessRequest) createConfig() RequestConfig {
	return RequestConfig{
		Method:       "POST",
		UrlSuffix:    ":7770/api/v1.0/aiserverless/create",
		ResourceName: "aiserverless",
		Headers:      nil,
	}
}
