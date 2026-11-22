package client

type StorageResourceMin struct {
    Name      string `json:"name"`
    Size      int    `json:"size"`
    Private   bool   `json:"private"`
    Shareable bool   `json:"shareable"`
    Bootable  bool   `json:"bootable"`
    Readonly  bool   `json:"readonly"`
}

type StorageResource struct {
    CreatorUUID      string   `json:"creatorID"`
    Name             string   `json:"name"`
    Size             int      `json:"size"`
    Private          bool     `json:"private"`
    Shareable        bool     `json:"shareable"`
    Bootable         bool     `json:"bootable"`
    Clonable         *bool    `json:"clonable,omitempty"`
    Algorithm        *string  `json:"alg,omitempty"`
    Format           *string  `json:"format,omitempty"`
    Bus              *string  `json:"bus,omitempty"`
    FS               *string  `json:"fs,omitempty"`
    HardwareDevice   *string  `json:"hw_device,omitempty"`
    Tags             []string `json:"tags,omitempty"`
    Region           string   `json:"region"`
    Readonly         bool     `json:"readonly"`
    Priority         *int     `json:"priority,omitempty"`
    BillingFrequency string   `json:"billing_frequency,omitempty"`
}

type CanCreateStorageRequest struct {
    Size             int                        `json:"size"`
    Region           string                     `json:"region"`
    BillingFrequency string                     `json:"billing_frequency,omitempty"`
    Benchmark        CanCreateStorageBenchmark  `json:"benchmark"`
    HardwareDevice   *string                    `json:"hw_device"`
    Filesystem       *string                    `json:"fs"`
    Ceph             *bool                      `json:"ceph"`
}

type CanCreateStorageBenchmark struct {
    ReadIOPS         int `json:"read_iops"`
    WriteIOPS        int `json:"write_iops"`
    ReadMBBandwidth  int `json:"read_MB_bw"`
    WriteMBBandwidth int `json:"write_MB_bw"`
}

type StorageResourceResponseMin struct {
    CreatorUUID string   `json:"creatorID"`
    Name        string   `json:"name"`
    Size        float64  `json:"size"`
    Private     bool     `json:"private"`
    Shareable   bool     `json:"shareable"`
    Bootable    bool     `json:"bootable"`
    Readonly    bool     `json:"readonly"`
    Clonable    bool     `json:"clonable"`
    Priority    int      `json:"priority"`
    Tags        []string `json:"tags"`
    UUID        string   `json:"volumeID"`
    LastUpdated string   `json:"lastUpdated"`
    Region      string   `json:"region"`
}

type StorageResourceResponse struct {
    CreatorUUID string   `json:"creatorID"`
    Name        string   `json:"name"`
    Size        float64  `json:"size"`
    Private     bool     `json:"private"`
    Shareable   bool     `json:"shareable"`
    Bootable    bool     `json:"bootable"`
    Readonly    bool     `json:"readonly"`
    Clonable    bool     `json:"clonable"`
    Priority    int      `json:"priority"`
    Tags        []string `json:"tags"`
    UUID        string   `json:"volumeID"`
    LastUpdated string   `json:"lastUpdated"`
    Region      string   `json:"region"`
    BillingUUID string   `json:"billing_uuid"`
}
