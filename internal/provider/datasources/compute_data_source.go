package datasources

import (
	"context"

	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &computeDataSource{}
	_ datasource.DataSourceWithConfigure = &computeDataSource{}
)

func NewComputeDataSource() datasource.DataSource {
	return &computeDataSource{}
}

type computeDataSource struct {
	common.BaseResource
}

type computeDataSourceModel struct {
	UUID             types.String                   `tfsdk:"uuid"`
	BillingUUID      types.String                   `tfsdk:"billing_uuid"`
	VMName           types.String                   `tfsdk:"vm_name"`
	Region           types.String                   `tfsdk:"region"`
	Tags             types.String                   `tfsdk:"tags"`
	Status           types.String                   `tfsdk:"status"`
	CPUSlots         types.Int32                    `tfsdk:"cpu_slots"`
	AllowSMT         types.Bool                     `tfsdk:"allow_smt"`
	MaxOverprovision types.Int32                    `tfsdk:"max_overprovision"`
	Arch             types.List                     `tfsdk:"arch"`
	Flags            types.List                     `tfsdk:"flags"`
	RAMSizeMB        types.Int32                    `tfsdk:"ram_size_mb"`
	RequireECC       types.Bool                     `tfsdk:"require_ecc"`
	OSFamily         types.String                   `tfsdk:"os_family"`
	OSFlavour        types.String                   `tfsdk:"os_flavour"`
	IPv4             types.String                   `tfsdk:"ipv4"`
	Volumes          []computeDataSourceVolumeModel `tfsdk:"volumes"`
}

type computeDataSourceVolumeModel struct {
	Name       types.String `tfsdk:"name"`
	Size       types.Int64  `tfsdk:"size"`
	VolumeUUID types.String `tfsdk:"volume_uuid"`
	Private    types.Bool   `tfsdk:"private"`
	Shareable  types.Bool   `tfsdk:"shareable"`
	Bootable   types.Bool   `tfsdk:"bootable"`
	Readonly   types.Bool   `tfsdk:"readonly"`
	Region     types.String `tfsdk:"region"`
}

func (d *computeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute"
}

func (d *computeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ids := identifierExactlyOneOf("uuid", "billing_uuid", "vm_name")
	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento virtual machine on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento virtual machine on the cloud selected by `host`. There is no Get-by-id for compute; the provider lists running VMs and matches exactly one of `uuid`, `billing_uuid`, or `vm_name`.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "VM uniqueID. Exactly one of uuid, billing_uuid, or vm_name must be set.",
				Validators:  ids,
			},
			"billing_uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Portal billing UUID of the VM.",
				Validators:  ids,
			},
			"vm_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "VM name. Not unique; use uuid when more than one match exists.",
				Validators:  ids,
			},
			"region":            schema.StringAttribute{Optional: true, Computed: true, Description: "Optional region filter, also populated from the API."},
			"tags":              schema.StringAttribute{Computed: true, Description: "Colon-separated VM tags."},
			"status":            schema.StringAttribute{Computed: true, Description: "Portal billing status, when billing_uuid is available."},
			"cpu_slots":         schema.Int32Attribute{Computed: true, Description: "Allocated CPU slots."},
			"allow_smt":         schema.BoolAttribute{Computed: true, Description: "Whether SMT is allowed."},
			"max_overprovision": schema.Int32Attribute{Computed: true, Description: "CPU overprovision factor."},
			"arch":              schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "CPU architecture."},
			"flags":             schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "CPU flags."},
			"ram_size_mb":       schema.Int32Attribute{Computed: true, Description: "Memory capacity in MB (API ramsize in GB multiplied by 1024)."},
			"require_ecc":       schema.BoolAttribute{Computed: true, Description: "Whether ECC memory is required."},
			"os_family":         schema.StringAttribute{Computed: true, Description: "OS family."},
			"os_flavour":        schema.StringAttribute{Computed: true, Description: "OS flavour."},
			"ipv4":              schema.StringAttribute{Computed: true, Description: "Primary IPv4 address when present."},
			"volumes": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Volumes attached to the VM.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":        schema.StringAttribute{Computed: true, Description: "Volume name."},
						"size":        schema.Int64Attribute{Computed: true, Description: "Volume size as returned by the API."},
						"volume_uuid": schema.StringAttribute{Computed: true, Description: "Volume UUID."},
						"private":     schema.BoolAttribute{Computed: true, Description: "Whether the volume is private."},
						"shareable":   schema.BoolAttribute{Computed: true, Description: "Whether the volume is shareable."},
						"bootable":    schema.BoolAttribute{Computed: true, Description: "Whether the volume is bootable."},
						"readonly":    schema.BoolAttribute{Computed: true, Description: "Whether the volume is read-only."},
						"region":      schema.StringAttribute{Computed: true, Description: "Volume region."},
					},
				},
			},
		},
	}
}

func (d *computeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Client = configureDataSourceClient(req, resp)
}

func (d *computeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config computeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := optionalString(config.UUID)
	billingUUID := optionalString(config.BillingUUID)
	vmName := optionalString(config.VMName)
	region := optionalString(config.Region)

	resources, err := d.Client.ListComputeResources()
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Compute Resources", "Could not list compute resources: "+err.Error())
		return
	}
	resources = filterByRegion(resources, region, func(item client.ComputeListItem) string { return item.Region })

	var matches []client.ComputeListItem
	for _, item := range resources {
		if uuid != "" && (item.UUID == uuid || item.ServiceUUID == uuid) {
			matches = append(matches, item)
			continue
		}
		if billingUUID != "" && item.BillingUUID == billingUUID {
			matches = append(matches, item)
			continue
		}
		if vmName != "" && item.Requirements.VMName == vmName {
			matches = append(matches, item)
		}
	}

	found, diags := requireOneMatch(matches, "elemento_compute")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.UUID = types.StringValue(found.UUID)
	config.BillingUUID = types.StringValue(found.BillingUUID)
	config.VMName = types.StringValue(found.Requirements.VMName)
	config.Region = types.StringValue(found.Region)
	config.Tags = types.StringValue(found.Tags)
	config.CPUSlots = types.Int32Value(int32(found.Requirements.Slots))
	config.AllowSMT = types.BoolValue(found.Requirements.AllowSMT)
	config.MaxOverprovision = types.Int32Value(int32(found.Requirements.Overprovision))
	config.RAMSizeMB = types.Int32Value(int32(found.Requirements.RamSize * 1024))
	config.RequireECC = types.BoolValue(found.Requirements.RequireECC)
	config.OSFamily = types.StringValue(found.Requirements.OSFamily)
	config.OSFlavour = types.StringValue(found.Requirements.OSFlavour)
	if found.Requirements.NetworkConfig.IPv4Address != "" {
		config.IPv4 = types.StringValue(found.Requirements.NetworkConfig.IPv4Address)
	} else {
		config.IPv4 = types.StringNull()
	}

	arch, archDiags := types.ListValueFrom(ctx, types.StringType, []string{found.Requirements.Arch})
	resp.Diagnostics.Append(archDiags...)
	config.Arch = arch

	flags, flagDiags := types.ListValueFrom(ctx, types.StringType, found.Requirements.Flags)
	resp.Diagnostics.Append(flagDiags...)
	config.Flags = flags

	var volumes []computeDataSourceVolumeModel
	for _, volume := range found.Requirements.Volumes {
		volumes = append(volumes, computeDataSourceVolumeModel{
			Name:       types.StringValue(volume.Name),
			Size:       types.Int64Value(int64(volume.Size)),
			VolumeUUID: types.StringValue(volume.UUID),
			Private:    types.BoolValue(volume.Private),
			Shareable:  types.BoolValue(volume.Shareable),
			Bootable:   types.BoolValue(volume.Bootable),
			Readonly:   types.BoolValue(volume.Readonly),
			Region:     types.StringValue(volume.Region),
		})
	}
	config.Volumes = volumes

	status, statusDiags := resourceStatus(d.Client, found.BillingUUID)
	resp.Diagnostics.Append(statusDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Status = status

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
