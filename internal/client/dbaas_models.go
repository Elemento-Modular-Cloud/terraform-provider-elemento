package client

type DBAASRequest struct {
    Name             string    `json:"name"`
    Engine           string    `json:"engine"`
    BackupTime       *string   `json:"backup_time,omitempty"`
    DiskSizeGB       int64     `json:"disk_size"`
    Password         string    `json:"password"`
    NodesNumber      int32     `json:"nodes_number"`
    Region           string    `json:"region"`
    BillingFrequency *string   `json:"billing_frequency,omitempty"`
}

type DBAASResponse struct {
    CreatorUUID   string    `json:"client_uuid"`
    Name          string    `json:"name"`
    Engine        string    `json:"engine"`
    BackupTime    string    `json:"backup_time"`
    DiskSizeGB    int64     `json:"disk_size"`
    NodesNumber   int32     `json:"nodes_number"`
    CreationDate  string    `json:"creation_date"`
    Region        string    `json:"region"`
    UUID          string    `json:"service_uuid"`
    BillingUUID   string    `json:"billing_uuid"`
}

type DBAASCredentials struct {
    DatabaseHost     string  `kv:"DB_HOST"`
    DatabasePort     string  `kv:"DB_PORT"`
    DatabaseUser     string  `kv:"DB_USER"`
    DatabasePassword string  `kv:"DB_PASSWORD"`
    DatabaseName     string  `kv:"DB_NAME"`
    DatabaseSSLMode  string  `kv:"DB_SSL_MODE"`
    PublicURI        string  `kv:"PUBLIC_URI"`
    ServiceUUID      string  `kv:"SERVICE_UUID"`
}
