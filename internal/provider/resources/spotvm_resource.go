package resources

import (
	"context"
	"time"

	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &spotVMResource{}
	_ resource.ResourceWithConfigure = &spotVMResource{}
)

func NewSpotVMResource() resource.Resource {
	return &spotVMResource{}
}

type spotVMResource struct {
	common.BaseResource
}

func (r *spotVMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spotvm"
}

func (r *spotVMResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	volumeAttributes := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name of the volume.",
		},
		"size_gb": schema.Int64Attribute{
			Required:    true,
			Description: "The size of the volume, in GB.",
		},
	}

	attributes["vm_name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the spotvm resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}

	attributes["cpu"] = schema.SingleNestedAttribute{
		Required:    true,
		Description: "The elemento spotvm cpu requirements",
		Attributes: map[string]schema.Attribute{
			"slots": schema.Int32Attribute{
				Description: "The number of CPU slots required for the spotvm resource.",
				Required:    true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"shared_cores": schema.BoolAttribute{
				Required:    true,
				Description: "Whether shared cores are allowed for the CPU.",
			},
			"arch": schema.ListAttribute{
				Required:    true,
				Description: "The list of alternative CPU architectures required for the spotvm resource. Valid values are 'X86_64' and 'arm64'.",
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							"X86_64",
							"arm64",
						),
					),
				},
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}

	attributes["mem"] = schema.SingleNestedAttribute{
		Required:    true,
		Description: "The elemento spotvm memory requirements",
		Attributes: map[string]schema.Attribute{
			"capacity_mb": schema.Int32Attribute{
				Required:    true,
				Description: "The memory capacity (in MB) required for the spotvm resource.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}

	attributes["os"] = schema.SingleNestedAttribute{
		Required:    true,
		Description: "The elemento spotvm miscellaneous requirements",
		Attributes: map[string]schema.Attribute{
			"family": schema.StringAttribute{
				Description: "The operating system family required for the spotvm resource. Valid values are 'linux', 'windows', and 'macos'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"linux",
						"windows",
						"macos",
					),
				},
			},
			"flavour": schema.StringAttribute{
				Description: "The operating system flavour required for the spotvm resource.",
				Required:    true, // TODO: os_flavour validator based on iso-templates?
			},
			"version": schema.StringAttribute{
				Description: "The operating system version required for the spotvm resource.",
				Optional:    true,
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}

	attributes["storage"] = schema.SingleNestedAttribute{
		Required:    true,
		Description: "The storage required for the spotvm resource.",
		Attributes: map[string]schema.Attribute{
			"boot": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The boot volume required for the spotvm storage.",
				Attributes:  volumeAttributes,
			},
			"data": schema.ListNestedAttribute{
				Required:    true,
				Description: "The list of optional data volumes required for the spotvm resource.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: volumeAttributes,
				},
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}

	// // TODO: Handle PCI
	// attributes["pci"] = schema.SingleNestedAttribute{
	// 	Required:    true,
	// 	Description: "The elemento spotvm PCI requirements",
	// 	Attributes: map[string]schema.Attribute{
	// 		"devices": schema.ListAttribute{
	// 			Required:    true,
	// 			Description: "A list of PCI device IDs required for the spotvm resource.",
	// 			ElementType: types.StringType,
	// 		},
	// 	},
	// 	PlanModifiers: []planmodifier.Object {
	// 		objectplanmodifier.RequiresReplace(),
	// 	},
	// }

	attributes["networks"] = schema.ListNestedAttribute{
		Computed:    true,
		Description: "The list of networks associated with the spotvm resource.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"kind": schema.StringAttribute{
					Computed:    true,
					Description: "The network kind.",
				},
				"ipv4": schema.StringAttribute{
					Computed:    true,
					Description: "The IPv4 address of the network.",
				},
				"mac": schema.StringAttribute{
					Computed:    true,
					Description: "The MAC address of the network.",
				},
			},
		},
	}

	attributes["auth"] = schema.SingleNestedAttribute{
		Required: true,
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"ssh_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}

	attributes["tags"] = schema.ListAttribute{
		Computed:    true,
		Optional:    true,
		Description: "The tags associated with the spotvm resource, assigned by the Elemento API after creation.",
		ElementType: types.StringType,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.ValueStringsAre(
				stringvalidator.OneOf(
					"X86_64",
					"arm64",
				),
			),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		Default: listdefault.StaticValue(
			types.ListValueMust(
				types.StringType,
				[]attr.Value{},
			),
		),
	}

	attributes["cloudinit_b64"] = schema.StringAttribute{
		Required:    true,
		Description: "The cloudinit file for the spotvm, base64-encoded.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *spotVMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.SpotVMResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cpuArch []string
	diags = plan.CPU.Arch.ElementsAs(ctx, &cpuArch, false)
	resp.Diagnostics.Append(diags...)

	var tags []string
	diags = plan.Tags.ElementsAs(ctx, &tags, false)
	resp.Diagnostics.Append(diags...)

	dataVolumes := []client.SpotVMVolume{}
	for _, volume := range plan.Storage.Data {
		dataVolumes = append(dataVolumes, client.SpotVMVolume{
			Name:   volume.Name.ValueString(),
			SizeGB: float64(volume.SizeGB.ValueInt64()),
		})
	}

	if resp.Diagnostics.HasError() {
		return
	}

	spotVMResource := client.SpotVMRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		VMName: plan.VMName.ValueString(),
		CPU: client.SpotVMCPU{
			Arch:        cpuArch,
			Slots:       int(plan.CPU.Slots.ValueInt32()),
			SharedCores: plan.CPU.SharedCores.ValueBool(),
		},
		Mem: client.SpotVMMemory{
			CapacityMB: int(plan.Mem.CapacityMB.ValueInt32()),
		},
		OS: client.SpotVMOS{
			Family:  plan.OS.Family.ValueString(),
			Flavour: plan.OS.Flavour.ValueString(),
			Version: plan.OS.Version.ValueString(),
		},
		Storage: client.SpotVMStorage{
			Boot: client.SpotVMVolume{
				Name:   plan.Storage.Boot.Name.ValueString(),
				SizeGB: float64(plan.Storage.Boot.SizeGB.ValueInt64()),
			},
			Data: dataVolumes,
		},
		PCI: client.SpotVMPCI{
			Devices: []client.SpotVMPCIDev{},
		},
		Networks: []client.SpotVMNetwork{
			{Kind: "natted"},
		},
		Auth: client.SpotVMAuth{
			Username: plan.Authentication.Username.ValueString(),
			Password: plan.Authentication.Password.ValueString(),
			SSHKey:   plan.Authentication.SSHKey.ValueStringPointer(),
		},
		Tags:            tags,
		CloudInitBase64: plan.CloudInitB64.ValueString(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(spotVMResource)
	common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "SpotVM", err)

	response, err := r.Client.CreateResource(spotVMResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating SpotVM Resource",
			"Could not create spotvm resource, unexpected error: "+err.Error(),
		)
		return
	}

	switch res := response.(type) {
	case client.PublicCreateResponse:
		plan.BillingUUID = types.StringValue(res.BillingUUID)
		plan.PaymentURL = types.StringValue(res.PaymentURL)
		plan.UUID = types.StringNull()

		status, err := r.Client.GetResourceStatus(res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading SpotVM Resource Status",
				"Could not read spotvm resource status: "+err.Error(),
			)
			return
		}

		plan.Status = types.StringValue(*status)
	case client.PrivateCreateResponse:
		plan.BillingUUID = types.StringNull()
		plan.PaymentURL = types.StringNull()
		plan.UUID = types.StringValue(res.ServiceUUID)
		plan.Status = types.StringValue(res.Status)
	}

	plan.Networks = types.ListNull(common.SpotVMNetworkObjectType())

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *spotVMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.SpotVMResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

    diags = model.Read(ctx, r.Client, common.RemoveResource(resp))
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
    	return
    }
    model.PaymentURL = types.StringNull()

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

func (r *spotVMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *spotVMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.SpotVMResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.DeleteSpotVMResource(state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SpotVM Resource",
			"Could not delete spotvm resoruce, unexpected error: "+err.Error(),
		)
		return
	}
}
