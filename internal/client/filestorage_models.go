package client

type FileStorageRequest struct {
	BaseRequest
	Name     string `json:"name"`
	SizeGB   int64  `json:"size_gb"`
	IsPublic bool   `json:"is_public"`
}

type FileStorageResponse struct {
	baseResponse
	Name    string  `json:"name"`
	SizeGB  int64   `json:"size_gb"`
	UsedGB  float64 `json:"used_gb"`
	DNSName string  `json:"dns_name"`
	State   string  `json:"state"`
}

type FileStorageCredentials struct {
	NFSHost      string `kv:"NFS_HOST"`
	NFSPath      string `kv:"NFS_PATH"`
	NFSVersion   string `kv:"NFS_VERSION"`
	MountCommand string `kv:"MOUNT_COMMAND"`
	ServiceUUID  string `kv:"SERVICE_UUID"`
}
