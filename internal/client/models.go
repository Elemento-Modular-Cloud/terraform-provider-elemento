package client

type ComputeResource struct {
    Requirements     ComputeRequirements  `json:"req"`
    VMName           string               `json:"vm_name"`
    Region           string               `json:"region"`
    CloudInit        CloudInitConfig      `json:"cloudinit"`
    AdditionalInfo   AdditionalInfoConfig `json:"additional_info"`
    BillingFrequency string               `json:"billing_frequency"`
    Volumes          []StorageVolume      `json:"volumes"`
}

type ComputeRequirements struct {
    CPU    ComputeCPU    `json:"cpu"`
    Memory ComputeMemory `json:"mem"`
    PCI    ComputePCI    `json:"pci"`
    Misc   ComputeMisc   `json:"misc"`
}

type ComputeCPU struct {
    SMTEnabled       bool     `json:"smtEnabled"`
    MinFrequency     float64  `json:"min_frequency"`
    Arch             string   `json:"arch"`
    Flags            []string `json:"flags"`
    Slots            int      `json:"slots"`
    AllowSMT         bool     `json:"allowSMT"`
    FullPhysical     bool     `json:"fullPhysical"`
    MaxOverprovision int      `json:"maxOverprovision"`
}

type ComputeMemory struct {
    Capacity int    `json:"capacity"`
    RequireECC bool `json:"requireECC"`
}

type ComputePCI struct {
    Devices map[string]int `json:"devices"`
}

type ComputeMisc struct {
    OSFamily string  `json:"os_family"`
    OSFlavour string `json:"os_flavour"`
}

type CloudInitConfig struct {
    Template  string            `json:"template"`
    Variables map[string]string `json:"variables"`
}

type AdditionalInfoConfig struct {
    Service string `json:"service"`
}

type StorageVolume struct {
    Name      string `json:"name"`
    Size      int    `json:"size"`
    Private   bool   `json:"private"`
    Shareable bool   `json:"shareable"`
    Bootable  bool   `json:"bootable"`
    Readonly  bool   `json:"readonly"`
}


type ComputeListItem struct {
    UUID             string             `json:"uniqueID"`
    Requirements     ComputeReqResponse `json:"req_json"`
    Region           string             `json:"region"`
    BillingUUID      string             `json:"billing_uuid"`
    States           string             `json:"states"`
    ServiceUUID      string             `json:"service_uuid"`
    Meson            MesonInfo          `json:"meson"`
    XML              string             `json:"xml"`
}

type ComputeReqResponse struct {
    Slots                int                     `json:"slots"`
    Overprovision        int                     `json:"overprovision"`
    AllowSMT             bool                    `json:"allowSMT"`
    Arch                 string                  `json:"arch"`
    Flags                []string                `json:"flags"`
    RamSize              float64                 `json:"ramsize"`
    RequireECC           bool                    `json:"reqECC"`
    Volumes              []StorageVolumeResponse `json:"volumes"`
    PciDevices           []string                `json:"pcidevs"`
    NetworkDevices       []string                `json:"netdevs"`
    OSFamily             string                  `json:"os_family"`
    OSFlavour            string                  `json:"os_flavour"`
    VMName               string                  `json:"vm_name"`
    CreationDate         string                  `json:"creation_date"`
    NetworkConfig        NetworkConfig           `json:"network_config"`
    Networks             []Network               `json:"networks"`
    PrivateNetworkConfig []PrivateNetworkConfig  `json:"private_network_config"`
}

type MesonInfo struct {
    Provider string   `json:"provider"`
    Price    struct{} `json:"price"`
}

type StorageVolumeResponse struct {
    CreatorUUID string   `json:"creatorID"`
    Name        string   `json:"name"`
    Size        int      `json:"size"`
    Private     bool     `json:"private"`
    Shareable   bool     `json:"shareable"`
    Bootable    bool     `json:"bootable"`
    Readonly    bool     `json:"readonly"`
    Clonable    bool     `json:"clonable"`
    Priority    int      `json:"priority"`
    Tags        []string `json:"tags"`
    UUID        string   `json:"volume_uuid"`
    LastUpdated string   `json:"lastUpdated"`
    Region      string   `json:"region"`
}

type NetworkConfig struct {
    DomDisplay        DomDisplay `json:"dom_display"`
    Interface         string     `json:"interface"`
    ReachableFromHost bool       `json:"is_reachable_from_host"`
    Model             string     `json:"model"`
    Name              string     `json:"name"`
    Source            string     `json:"source"`
    Type              string     `json:"type"`
    IPv4Address       string     `json:"ipv4"`
    IPv6Address       string     `json:"ipv6"`
    MACAddress        string     `json:"mac"`
}

type DomDisplay struct {
    Port     int    `json:"port"`
    Protocol string `json:"protocol"`
}

type PrivateNetworkConfig struct {
    IPv4Address       string `json:"ipv4"`
    IPv6Address       string `json:"ipv6"`
    MACAddress        string `json:"mac"`
}

type Network struct {
}
