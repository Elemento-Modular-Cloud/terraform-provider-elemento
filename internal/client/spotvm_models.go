package client

type SpotVMResource struct {
	VMName string `json:"vm_name"`
	CPU    struct {
		Slots       int      `json:"slots"`
		SharedCores bool     `json:"shared_cores"`
		Arch        []string `json:"arch"`
	} `json:"cpu"`
	Mem struct {
		CapacityMB int `json:"capacity_mb"`
	} `json:"mem"`
	OS struct {
		Family  string `json:"family"`
		Flavour string `json:"flavour"`
		Version string `json:"version"`
	} `json:"os"`
	Storage struct {
		Boot SpotVMVolume   `json:"boot"`
		Data []SpotVMVolume `json:"data"`
	} `json:"storage"`
	PCI struct {
		Devices []SpotVMPCIDev `json:"devices"`
	} `json:"pci"`
	Networks []SpotVMNetwork `json:"networks"`
	Auth struct {
		Username string `json:"username"`
		Password string `json:"password"`
		SSHKey   string `json:"ssh_key"`
	} `json:"auth"`
	Tags            []string `json:"tags"`
	CloudInitBase64 string   `json:"cloud_init_b64"`
	region          string
}

type SpotVMVolume struct {
	UUID   *string `json:"volume_uuid,omitempty"`
	Name   string `json:"name"`
	SizeGB int    `json:"size_gb"`
}

type SpotVMPCIDev struct {
	Vendor   string `json:"vendor"`
	Model    string `json:"model"`
	Quantity int    `json:"quantity"`
}

type SpotVMNetwork struct {
	Kind string `json:"kind"`
}

func (r *ComputeResource) toSpotVM() SpotVMResource {
	spot := SpotVMResource{}

	spot.VMName = r.VMName

	cpu := r.Requirements.CPU
	spot.CPU.Slots = cpu.Slots
	spot.CPU.Arch = cpu.Arch
	spot.CPU.SharedCores = !cpu.FullPhysical

	spot.Mem.CapacityMB = r.Requirements.Memory.Capacity

	misc := r.Requirements.Misc
	spot.OS.Family = misc.OSFamily
	spot.OS.Flavour = misc.OSFlavour
	// Version has no equivalent field in ComputeResource; left as zero value.

	// Storage: first bootable volume -> Boot, remaining -> Data
	dataVolumes := make([]SpotVMVolume, 0, len(r.Volumes))
	bootAssigned := false
	for _, v := range r.Volumes {
		sv := SpotVMVolume{
			Name:   v.Name,
			SizeGB: v.Size,
		}
		if v.Bootable && !bootAssigned {
			spot.Storage.Boot = sv
			bootAssigned = true
		} else {
			dataVolumes = append(dataVolumes, sv)
		}
	}
	spot.Storage.Data = dataVolumes

	spot.PCI = struct {
		Devices []SpotVMPCIDev `json:"devices"`
	}{
		Devices: []SpotVMPCIDev{},
	}

	spot.Networks = []SpotVMNetwork {
		{Kind: "natted"},
	}

	if auth := r.Authentication; auth != nil {
		spot.Auth.Username = auth.Username
		spot.Auth.Password = auth.Password
		if auth.SSHKey != nil {
			spot.Auth.SSHKey = *auth.SSHKey
		}
	}

	// TODO: CloudInit

	spot.region = r.Region

	return spot
}
