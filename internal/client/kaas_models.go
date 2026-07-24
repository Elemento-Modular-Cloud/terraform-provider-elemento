package client

type KAASRequest struct {
    Name                   string         `json:"name"`
    Region                 string         `json:"region"`
    NodePools              []KAASNodePool `json:"nodepools"`
    NodesSubnet            KaasNodeSubnet `json:"nodes_subnet"`
    UpdatePolicy           string         `json:"update_policy"`
    KubernetesVersion      string         `json:"kubernetes_version"`
    ControlPlaneHAEnabled  bool           `json:"controlplane_ha"`
    BillingFrequency       string         `json:"billing_frequency"`
}

type KAASNodePool struct {
    Name             string            `json:"pool_name"`
    AutoScale        bool              `json:"autoscale"`
    MinNodes         int32             `json:"min_nodes"`
    MaxNodes         int32             `json:"max_nodes"`
    SharedStorageGB  int64             `json:"shared_storage_gb"`
    Specs            KAASNodePoolSpecs `json:"specs"`
}

type KAASNodePoolSpecs struct {
    CPUCores          int       `json:"cpu_cores"`
    MemoryGB          int       `json:"ram_gb"`
    DiskSizeGB        int       `json:"storage_gb"`
    GPUs              int       `json:"gpu"`
}

type KaasNodeSubnet struct {
    DHCPEnabled bool   `json:"dhcp"`
    CIDR        string `json:"network"`
}

type KAASResponse struct {
    CreatorUUID   string    `json:"client_uuid"`
    Name          string    `json:"name"`
    ClusterName   string    `json:"cluster_name"`
    Status        string    `json:"status"`
    Endpoint      string    `json:"endpoint"`
    Region        string    `json:"location"`
    KubeConfig    string    `json:"kubeconfig"`
    Version       string    `json:"version"`
    NetworkCIDR   string    `json:"network_cidr"`
    UUID          string    `json:"service_uuid"`
    BillingUUID   string    `json:"billing_uuid"`
}
