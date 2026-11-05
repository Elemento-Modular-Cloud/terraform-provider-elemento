package resources

import (
	"context"
	"fmt"
	"time"

	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
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
	_ resource.Resource = &computeResource{}
	_ resource.ResourceWithConfigure = &computeResource{}
)

func NewComputeResource() resource.Resource {
	return &computeResource{}
}

type computeResource struct{
	common.BaseResource
}

type computeResourceModel struct {
	Requirements     computeRequirementsModel    `tfsdk:"requirements"`
	Authentication   *computeAuthenticationModel `tfsdk:"authentication"`
	Volumes          []storageResourceModel      `tfsdk:"volumes"`
	VMName           types.String                `tfsdk:"vm_name"`
    Region           types.String                `tfsdk:"region"`
    CloudInit        computeCloudInitModel       `tfsdk:"cloudinit"`
    AdditionalInfo   computeAdditionalInfoModel  `tfsdk:"additional_info"`
    BillingFrequency types.String                `tfsdk:"billing_frequency"`
    BillingUUID      types.String                `tfsdk:"billing_uuid"`
    PaymentURL       types.String                `tfsdk:"payment_url"`
    LastUpdated      types.String                `tfsdk:"last_updated"`
    UUID             types.String                `tfsdk:"uuid"`
    Tags             types.String                `tfsdk:"tags"`
    Status           types.String                `tfsdk:"status"`
}

type computeRequirementsModel struct {
	CPU    computeCPUModel    `tfsdk:"cpu"`
	Memory computeMemoryModel `tfsdk:"memory"`
	PCI    computePCIModel 	  `tfsdk:"pci"`
	Misc   computeMiscModel   `tfsdk:"misc"`
}

type computeCPUModel struct {
	SMTEnabled 		 types.Bool    `tfsdk:"smt_enabled"`
	MinFrequency 	 types.Float64 `tfsdk:"min_frequency"`
	Arch 			 types.List    `tfsdk:"arch"`
	Flags 			 types.List    `tfsdk:"flags"`
	Slots 			 types.Int32   `tfsdk:"slots"`
	AllowSMT 		 types.Bool    `tfsdk:"allow_smt"`
	FullPhysical 	 types.Bool    `tfsdk:"full_physical"`
	MaxOverprovision types.Int32   `tfsdk:"max_overprovision"`
}

type computeMemoryModel struct {
	Capacity   types.Int32 `tfsdk:"capacity"`
	RequireECC types.Bool  `tfsdk:"require_ecc"`
}

type computePCIModel struct {
	Devices types.Map `tfsdk:"devices"`
}

type computeMiscModel struct {
	OSFamily  types.String `tfsdk:"os_family"`
	OSFlavour types.String `tfsdk:"os_flavour"`
}

type computeCloudInitModel struct {
    Template  types.String `tfsdk:"template"`
    Variables types.Map    `tfsdk:"variables"`
}

type computeAuthenticationModel struct {
    Username types.String `tfsdk:"username"`
    Password types.String `tfsdk:"password"`
    SSHKey   types.String `tfsdk:"ssh_key"`
}

type computeAdditionalInfoModel struct {
    Service types.String `tfsdk:"service"`
}

func (r *computeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute"
}

func (r *computeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The elemento compute resource",
		Attributes: map[string]schema.Attribute{
			"requirements": schema.SingleNestedAttribute{
				Required: true,
				Description: "The elemento compute requirements",
				Attributes: map[string]schema.Attribute{
					"cpu": schema.SingleNestedAttribute{
						Required: true,
						Description: "The elemento compute cpu requirements",
						Attributes: map[string]schema.Attribute{
							"smt_enabled": schema.BoolAttribute{ // TODO: is this a dupe?
								Required: true,
								Description: "Whether SMT (Simultaneous Multithreading) is enabled for the CPU.",
							},
							"min_frequency": schema.Float64Attribute{
								Optional: true,
								Description: "The minimum frequency (in GHz) required for the CPU.",
								Validators: []validator.Float64{
									float64validator.AtLeast(0),
								},
							},
							"arch": schema.ListAttribute{
							    Required: true,
								Description: "The list of alternative CPU architectures required for the compute resource. Valid values are 'X86_64' and 'arm64'.",
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
							"flags": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "The list of CPU flags required for the compute resource.",
								Optional:    true,
								Computed:    true,
								Default:     listdefault.StaticValue(
									types.ListValueMust(
										types.StringType,
										[]attr.Value{},
									),
								),
							},
							"slots": schema.Int32Attribute{
			                    Description: "The number of CPU slots required for the compute resource.",
								Required: true,
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
								},
							},
							"allow_smt": schema.BoolAttribute{
								Required: true,
								Description: "Whether SMT (Simultaneous Multithreading) is allowed for the CPU.",
							},
							"full_physical": schema.BoolAttribute{ // TODO: is this a dupe?
								Required: true,
								Description: "Whether the CPU slots should be a full physical CPU core (no hyperthreading).",
							},
							"max_overprovision": schema.Int32Attribute{
								Required: true,
								Description: "The maximum overprovisioning factor for the CPU slots.",
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
								},
							},
						},
					},
					"memory": schema.SingleNestedAttribute{
						Required: true,
						Description: "The elemento compute memory requirements",
						Attributes: map[string]schema.Attribute{
							"capacity": schema.Int32Attribute{
								Required: true,
								Description: "The memory capacity (in MB) required for the compute resource.",
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
								},
							},
							"require_ecc": schema.BoolAttribute{
								Required: true,
								Description: "Whether ECC (Error-Correcting Code) memory is required for the compute resource.",
							},
						},
					},
					"pci": schema.SingleNestedAttribute{
						Required: true,
						Description: "The elemento compute PCI requirements",
						Attributes: map[string]schema.Attribute{
							"devices": schema.MapAttribute{
								Required: true,
								Description: "A map of PCI device names to the number of devices required for the compute resource.",
								ElementType: types.Int32Type,
								Validators: []validator.Map{
									mapvalidator.ValueInt32sAre(int32validator.AtLeast(0)),
								},
							},
						},
					},
					"misc": schema.SingleNestedAttribute{
						Required: true,
						Description: "The elemento compute miscellaneous requirements",
						Attributes: map[string]schema.Attribute{
							"os_family": schema.StringAttribute{
			                    Description: "The operating system family required for the compute resource. Valid values are 'linux', 'windows', and 'macos'.",
								Optional: true,
								Validators: []validator.String{
									stringvalidator.OneOf(
										"linux",
										"windows",
										"macos",
									),
								},
							},
							"os_flavour": schema.StringAttribute{
							    Description: "The operating system flavour required for the compute resource.",
								Optional: true, // TODO: os_flavour validator based on iso-templates?
							},
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
				    objectplanmodifier.RequiresReplace(),
				},
			},
			"authentication": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Required: true,
					},
					"password": schema.StringAttribute{
						Required: true,
						Sensitive: true,
					},
					"ssh_key": schema.StringAttribute{
						Optional: true,
						Sensitive: true,
					},
				},
				PlanModifiers: []planmodifier.Object{
				    objectplanmodifier.RequiresReplace(),
				},
			},
			"volumes": schema.ListNestedAttribute{
				Required: true,
				Description: "The list of storage volumes required for the compute resource.",
				NestedObject: schema.NestedAttributeObject{
                    Attributes: volumeAttributes(),
				},
				PlanModifiers: []planmodifier.List{
                    listplanmodifier.RequiresReplace(),
				},
			},
			"vm_name": schema.StringAttribute{
				Required: true,
				Description: "The name of the virtual machine to be created.",
				PlanModifiers: []planmodifier.String{
				    stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required: true,
				Description: "The region in which the compute resource should be created.",
				PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
			},
			"cloudinit": schema.SingleNestedAttribute{
			    Optional: true,
				Description: "The cloud-init configuration for the compute resource.",
				Attributes: map[string]schema.Attribute{
					"template": schema.StringAttribute{
	                    Description: "The cloud-init template to be used for the compute resource.",
						Required: true,
					},
					"variables": schema.MapAttribute{
					    Description: "A map of variables to be used in the cloud-init template.",
						Optional: true,
						ElementType: types.StringType,
					},
				},
				PlanModifiers: []planmodifier.Object{
                    objectplanmodifier.RequiresReplace(),
				},
			},
			"additional_info": schema.SingleNestedAttribute{
				Required: true,
				Description: "Additional information for the compute resource.",
				Attributes: map[string]schema.Attribute{
					"service": schema.StringAttribute{
					    Description: "The service for which the compute resource is being created.",
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.Object{
                    objectplanmodifier.RequiresReplace(),
                },
			},
			"billing_frequency": schema.StringAttribute{
				Required: true,
				Description: "The frequency at which the compute resource should be billed. Valid values are 'day', 'week', 'month', and 'year'.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"day",
						"month",
						"year",
					),
				},
				PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
			},
			"billing_uuid": schema.StringAttribute{
                Computed: true,
                Description: "The billing UUID of the compute resource, assigned by the Elemento API after creation.",
            },
            "payment_url": schema.StringAttribute{
                Computed: true,
                Description: "The payment URL for the compute resource, assigned by the Elemento API after creation.",
            },
            "last_updated": schema.StringAttribute{
                Computed: true,
                Description: "The timestamp of the last update to the compute resource.",
            },
            "uuid": schema.StringAttribute{
                Computed: true,
                Description: "The UUID of the compute resource, assigned by the Elemento API after creation.",
            },
            "tags": schema.StringAttribute{
                Computed: true,
                Description: "The tags associated with the compute resource, assigned by the Elemento API after creation.",
            },
            "status": schema.StringAttribute{
                Computed: true,
                Description: "The status of the compute resource, assigned by the Elemento API after creation.",
            },
		},
	}
}

func (r *computeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ElementoClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *ElementoClient, got %T. Please report this issue to the provider developer.", req.ProviderData),
		)

		return
	}
	r.Client = client
}

func (r *computeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan computeResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var cpuArch []string
    diags = plan.Requirements.CPU.Arch.ElementsAs(ctx, &cpuArch, false)
    resp.Diagnostics.Append(diags...)

    var cpuFlags []string
    diags = plan.Requirements.CPU.Flags.ElementsAs(ctx, &cpuFlags, false)
    resp.Diagnostics.Append(diags...)

    var pciDevices map[string]int
    diags = plan.Requirements.PCI.Devices.ElementsAs(ctx, &pciDevices, false)
    resp.Diagnostics.Append(diags...)

    var cloudinitVariables map[string]string
    diags = plan.CloudInit.Variables.ElementsAs(ctx, &cloudinitVariables, false)
    resp.Diagnostics.Append(diags...)

    var volumes []client.StorageResourceMin
    for _, volume := range plan.Volumes {
        volumes = append(volumes, client.StorageResourceMin{
            Name:      volume.Name.ValueString(),
            Size:      int(volume.Size.ValueInt64()),
            Private:   volume.Private.ValueBool(),
            Shareable: volume.Shareable.ValueBool(),
            Bootable:  volume.Bootable.ValueBool(),
            Readonly:  volume.Readonly.ValueBool(),
        })
    }

    var authentication *client.ComputeAuthentication = nil
    if plan.Authentication != nil {
        authentication = &client.ComputeAuthentication{
            Username: plan.Authentication.Username.ValueString(),
            Password: plan.Authentication.Password.ValueString(),
            SSHKey:   plan.Authentication.SSHKey.ValueStringPointer(),
        }
    }

    if resp.Diagnostics.HasError() {
        return
    }

    computeResource := client.ComputeResource{
        Requirements: client.ComputeRequirements{
            CPU: client.ComputeCPU{
                SMTEnabled:       plan.Requirements.CPU.SMTEnabled.ValueBool(),
                MinFrequency:     plan.Requirements.CPU.MinFrequency.ValueFloat64(),
                Arch:             cpuArch,
                Flags:            cpuFlags,
                Slots:            int(plan.Requirements.CPU.Slots.ValueInt32()),
                AllowSMT:         plan.Requirements.CPU.AllowSMT.ValueBool(),
                FullPhysical:     plan.Requirements.CPU.FullPhysical.ValueBool(),
                MaxOverprovision: int(plan.Requirements.CPU.MaxOverprovision.ValueInt32()),
            },
            Memory: client.ComputeMemory{
                Capacity:   int(plan.Requirements.Memory.Capacity.ValueInt32()),
                RequireECC: plan.Requirements.Memory.RequireECC.ValueBool(),
            },
            PCI: client.ComputePCI{
                Devices: pciDevices,
            },
            Misc: client.ComputeMisc{
                OSFamily:  plan.Requirements.Misc.OSFamily.ValueString(),
                OSFlavour: plan.Requirements.Misc.OSFlavour.ValueString(),
            },
        },
        Authentication: authentication,
        VMName: plan.VMName.ValueString(),
        Region: plan.Region.ValueString(),
        CloudInit: client.CloudInitConfig{
            Template:  plan.CloudInit.Template.ValueString(),
            Variables: cloudinitVariables,
        },
        AdditionalInfo: client.AdditionalInfoConfig{
            Service: plan.AdditionalInfo.Service.ValueString(),
        },
        BillingFrequency: plan.BillingFrequency.ValueString(),
        Volumes: volumes,
    }

    canCreateResponse, err := r.Client.CanCreateComputeResource(computeResource)
	common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "Compute", err)

    response, err := r.Client.CreateComputeResource(computeResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating Compute Resource",
            "Could not create compute resource, unexpected error: " + err.Error(),
        )
        return
    }

    switch res := response.(type) {
        case client.PublicCreateResponse:
            plan.BillingUUID = types.StringValue(res.BillingUUID)
            plan.PaymentURL  = types.StringValue(res.PaymentURL)
            plan.UUID        = types.StringNull()

            status, err := r.Client.GetResourceStatus(res.BillingUUID)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Error Reading Compute Resource Status",
                    "Could not read compute resource status: "+err.Error(),
                )
                return
            }

            plan.Status = types.StringValue(*status)
        case client.PrivateCreateResponse:
            plan.BillingUUID = types.StringNull()
            plan.PaymentURL  = types.StringNull()
            plan.UUID        = types.StringValue(res.ServiceUUID)
            plan.Status      = types.StringValue(res.Status)
    }

    plan.Tags = types.StringNull()

    for volumeIndex, volume := range plan.Volumes {
        volume.Priority           = types.Int32Null()
        volume.Clonable           = types.BoolNull()
        volume.UUID               = types.StringNull()
        volume.LastUpdated        = types.StringNull()
        volume.BillingUUID        = types.StringNull()
        volume.CreatorUUID        = types.StringNull()
        volume.Region             = types.StringNull()
        volume.PaymentURL         = types.StringNull()
        volume.Tags               = types.ListValueMust(types.StringType, []attr.Value{})
        plan.Volumes[volumeIndex] = volume
    }

    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *computeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state computeResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    resources, err := r.Client.ListComputeResources()
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading Compute Resources",
            "Could not list compute resources: "+err.Error(),
        )
        return
    }

    var resource *client.ComputeListItem;
    for _, res := range resources {
        if res.BillingUUID == state.BillingUUID.ValueString() || res.UUID == state.UUID.ValueString() {
            resource = &res
            break
        }
    }

    billing_uuid := state.BillingUUID.ValueStringPointer()
    var status *string
    if billing_uuid != nil && *billing_uuid != "" {
        status, err = r.Client.GetResourceStatus(state.BillingUUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading Compute Resource Status",
                "Could not read compute resource status: "+err.Error(),
            )
            return
        }
    } else {
        status = nil
    }

    if resource == nil && (status == nil || *status == "deleted" || *status == "error" || *status == "canceled") {
        resp.State.RemoveResource(ctx)
        return
    }

    state.Requirements.CPU.Arch, diags = types.ListValueFrom(ctx, types.StringType, resource.Requirements.Arch)
    resp.Diagnostics.Append(diags...)

    state.Requirements.CPU.Flags, diags = types.ListValueFrom(ctx, types.StringType, resource.Requirements.Flags)
    resp.Diagnostics.Append(diags...)

    state.Requirements.CPU.Slots            = types.Int32Value(int32(resource.Requirements.Slots))
    state.Requirements.CPU.AllowSMT         = types.BoolValue(resource.Requirements.AllowSMT)
    state.Requirements.CPU.MaxOverprovision = types.Int32Value(int32(resource.Requirements.Overprovision))

    state.Requirements.Memory.Capacity   = types.Int32Value(int32(resource.Requirements.RamSize * 1024))
    state.Requirements.Memory.RequireECC = types.BoolValue(resource.Requirements.RequireECC)

    // TODO: PCI DEVICES

    state.Requirements.Misc.OSFamily  = types.StringValue(resource.Requirements.OSFamily)
    state.Requirements.Misc.OSFlavour = types.StringValue(resource.Requirements.OSFlavour)

    var volumes []storageResourceModel
    for _, volume := range resource.Requirements.Volumes {
        tags, diags := types.ListValueFrom(ctx, types.StringType, volume.Tags)
        resp.Diagnostics.Append(diags...)

        volumes = append(volumes, storageResourceModel{
            Name:        types.StringValue(volume.Name),
            Size:        types.Int64Value(int64(volume.Size)),
            Private:     types.BoolValue(volume.Private),
            Shareable:   types.BoolValue(volume.Shareable),
            Bootable:    types.BoolValue(volume.Bootable),
            Readonly:    types.BoolValue(volume.Readonly),
            Clonable:    types.BoolValue(volume.Clonable),
            Priority:    types.Int32Value(int32(volume.Priority)),
            UUID:        types.StringValue(volume.UUID),
            LastUpdated: types.StringValue(volume.LastUpdated),
            Tags:        tags,
            Region:      types.StringValue(volume.Region),
        })
    }
    state.Volumes = volumes

    state.Status      = types.StringPointerValue(status)
    state.VMName      = types.StringValue(resource.Requirements.VMName)
    state.Region      = types.StringValue(resource.Region)
    state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    state.UUID        = types.StringValue(resource.UUID)
    state.Tags        = types.StringValue(resource.Tags)
    state.PaymentURL  = types.StringNull()

    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *computeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *computeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state computeResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.Client.DeleteComputeResource(state.UUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Deleting Compute Resource",
            "Could not delete compute resoruce, unexpected error: " + err.Error(),
        )
        return
    }
}
