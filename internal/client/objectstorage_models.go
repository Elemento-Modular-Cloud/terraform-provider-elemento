package client

type ObjectStorageRequest struct {
	Name             string `json:"name"`
	SizeTB           int    `json:"purchasedTB"`
	Region           string `json:"region"`
	BillingFrequency string `json:"billing_frequency"`
}

type ObjectStorageResponse struct {
	ActiveStorage int32  `json:"active_storage"`
	Name          string `json:"name"`
	Region        string `json:"region"`
	UUID          string `json:"service_uuid"`
	Endpoint      string `json:"endpoint"`
	BillingUUID   string `json:"billing_uuid"`
}

type ObjectStorageCredentials struct {
	AccountName     string `kv:"ACCOUNT_NAME"`
	ContainerName   string `kv:"CONTAINER_NAME"`
	SasToken        string `kv:"SAS_TOKEN"`
	AzureCliCommand string `kv:"AZURE_CLI_COMMAND"`
}
