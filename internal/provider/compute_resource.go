package provider

import (
	"context"
	"fmt"
	"time"

	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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
	client *client.ElementoClient
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
	Arch 			 types.String  `tfsdk:"arch"`
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
		Attributes: map[string]schema.Attribute{
			"requirements": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"cpu": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"smt_enabled": schema.BoolAttribute{ // TODO: is this a dupe?
								Required: true,
							},
							"min_frequency": schema.Float64Attribute{
								Optional: true,
								Validators: []validator.Float64{
									float64validator.AtLeast(0),
								},
							},
							"arch": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf(
										"X86_64",
										"arm64",
									),
								},
							},
							"flags": schema.ListAttribute{
								ElementType: types.StringType,
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
								Required: true,
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
								},
							},
							"allow_smt": schema.BoolAttribute{
								Required: true,
							},
							"full_physical": schema.BoolAttribute{ // TODO: is this a dupe?
								Required: true,
							},
							"max_overprovision": schema.Int32Attribute{
								Required: true,
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
								},
							},
						},
					},
					"memory": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"capacity": schema.Int32Attribute{
								Required: true,
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
								},
							},
							"require_ecc": schema.BoolAttribute{
								Required: true,
							},
						},
					},
					"pci": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"devices": schema.MapAttribute{
								Required: true,
								ElementType: types.Int32Type,
								Validators: []validator.Map{
									mapvalidator.ValueInt32sAre(int32validator.AtLeast(0)),
								},
							},
						},
					},
					"misc": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"os_family": schema.StringAttribute{
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
				NestedObject: schema.NestedAttributeObject{
                    Attributes: volumeAttributes(),
				},
				PlanModifiers: []planmodifier.List{
                    listplanmodifier.RequiresReplace(),
				},
			},
			"vm_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
				    stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
			},
			"cloudinit": schema.SingleNestedAttribute{
			    Optional: true,
				Attributes: map[string]schema.Attribute{
					"template": schema.StringAttribute{
						Required: true,
					},
					"variables": schema.MapAttribute{
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
				Attributes: map[string]schema.Attribute{
					"service": schema.StringAttribute{
						Required: true,
					},
				},
				PlanModifiers: []planmodifier.Object{
                    objectplanmodifier.RequiresReplace(),
                },
			},
			"billing_frequency": schema.StringAttribute{
				Required: true,
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
            },
            "payment_url": schema.StringAttribute{
                Computed: true,
            },
            "last_updated": schema.StringAttribute{
                Computed: true,
            },
            "uuid": schema.StringAttribute{
                Computed: true,
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
	r.client = client
}

func (r *computeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan computeResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

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

    authentication := client.ComputeAuthentication{
        Username: plan.Authentication.Username.ValueString(),
        Password: plan.Authentication.Password.ValueString(),
        SSHKey:   plan.Authentication.SSHKey.ValueStringPointer(),
    }

    if resp.Diagnostics.HasError() {
        return
    }

    computeResource := client.ComputeResource{
        Requirements: client.ComputeRequirements{
            CPU: client.ComputeCPU{
                SMTEnabled:       plan.Requirements.CPU.SMTEnabled.ValueBool(),
                MinFrequency:     plan.Requirements.CPU.MinFrequency.ValueFloat64(),
                Arch:             plan.Requirements.CPU.Arch.ValueString(),
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
        Authentication: &authentication,
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

    response, err := r.client.CreateComputeResource(computeResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating compute resource",
            "Could not create compute resource, unexpected error: " + err.Error(),
        )
        return
    }

    plan.BillingUUID = types.StringValue(response.BillingUUID)
    plan.PaymentURL = types.StringValue(response.PaymentURL)
    plan.UUID = types.StringNull()

    for volumeIndex, volume := range plan.Volumes {
        volume.Priority = types.Int32Null()
        volume.Clonable = types.BoolNull()
        volume.UUID = types.StringNull()
        volume.LastUpdated = types.StringNull()
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

    resources, err := r.client.ListComputeResources(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading Compute Resources",
            "Could not list compute resources: "+err.Error(),
        )
        return
    }

    var resource *client.ComputeListItem;
    for _, res := range resources {
        if res.BillingUUID == state.BillingUUID.ValueString() {
            resource = &res
            break
        }
    }

    if resource == nil {
        resp.State.RemoveResource(ctx)
        return
    }

    state.Requirements.CPU.Arch = types.StringValue(resource.Requirements.Arch)
    state.Requirements.CPU.Flags, diags = types.ListValueFrom(ctx, types.StringType, resource.Requirements.Flags)
    resp.Diagnostics.Append(diags...)
    state.Requirements.CPU.Slots = types.Int32Value(int32(resource.Requirements.Slots))
    state.Requirements.CPU.AllowSMT = types.BoolValue(resource.Requirements.AllowSMT)
    state.Requirements.CPU.MaxOverprovision = types.Int32Value(int32(resource.Requirements.Overprovision))

    state.Requirements.Memory.Capacity = types.Int32Value(int32(resource.Requirements.RamSize * 1024))
    state.Requirements.Memory.RequireECC = types.BoolValue(resource.Requirements.RequireECC)

    // TODO: PCI DEVICES

    state.Requirements.Misc.OSFamily = types.StringValue(resource.Requirements.OSFamily)
    state.Requirements.Misc.OSFlavour = types.StringValue(resource.Requirements.OSFlavour)

    var volumes []storageResourceModel
    for _, volume := range resource.Requirements.Volumes {
        volumes = append(volumes, storageResourceModel{
            Name:      types.StringValue(volume.Name),
            Size:      types.Int64Value(int64(volume.Size)),
            Private:   types.BoolValue(volume.Private),
            Shareable: types.BoolValue(volume.Shareable),
            Bootable:  types.BoolValue(volume.Bootable),
            Readonly:  types.BoolValue(volume.Readonly),
            Clonable:  types.BoolValue(volume.Clonable),
            Priority:  types.Int32Value(int32(volume.Priority)),
            UUID: types.StringValue(volume.UUID),
            LastUpdated: types.StringValue(volume.LastUpdated),
        })
    }
    state.Volumes = volumes

    state.VMName = types.StringValue(resource.Requirements.VMName)
    state.Region = types.StringValue(resource.Region)
    state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    state.UUID = types.StringValue(resource.UUID)
    state.PaymentURL = types.StringNull()

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

    err := r.client.DeleteComputeResource(state.UUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Deleting Compute Resource",
            "Could not delete compute resoruce, unexpected error: " + err.Error(),
        )
        return
    }
}
