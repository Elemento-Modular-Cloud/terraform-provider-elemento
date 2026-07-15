package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
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
	client *ElementoClient
}

type computeResourceModel struct {
	Requirements     computeRequirementsModel   `tfsdk:"requirements"`
	Volumes          []storageVolumeModel       `tfsdk:"volumes"`
	VMName           types.String               `tfsdk:"vm_name"`
    Region           types.String               `tfsdk:"region"`
    CloudInit        computeCloudInitModel      `tfsdk:"cloudinit"`
    AdditionalInfo   computeAdditionalInfoModel `tfsdk:"additional_info"`
    BillingFrequency types.String               `tfsdk:"billing_frequency"`
    BillingUUID      types.String               `tfsdk:"billing_uuid"`
    PaymentURL       types.String               `tfsdk:"payment_url"`
    LastUpdated      types.String               `tfsdk:"last_updated"`
    UUID             types.String               `tfsdk:"vm_uuid"`
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
			},
			"volumes": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
                    Attributes: volumeAttributes(),
				},
			},
			"vm_name": schema.StringAttribute{
				Required: true,
			},
			"region": schema.StringAttribute{
				Required: true,
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
			},
			"additional_info": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"service": schema.StringAttribute{
						Required: true,
					},
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
            "vm_uuid": schema.StringAttribute{
                Computed: true,
            },
		},
	}
}

func (r *computeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ElementoClient)

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
    // Retrieve values from plan
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

    var volumes []StorageVolume
    for _, volume := range plan.Volumes {
        volumes = append(volumes, StorageVolume{
            Name:      volume.Name.ValueString(),
            Size:      int(volume.Size.ValueInt64()),
            Private:   volume.Private.ValueBool(),
            Shareable: volume.Shareable.ValueBool(),
            Bootable:  volume.Bootable.ValueBool(),
            Readonly:  volume.Readonly.ValueBool(),
        })
    }

    if resp.Diagnostics.HasError() {
        return
    }

    computeResource := ComputeResource{
        Requirements: ComputeRequirements{
            CPU: ComputeCPU{
                SMTEnabled:       plan.Requirements.CPU.SMTEnabled.ValueBool(),
                MinFrequency:     plan.Requirements.CPU.MinFrequency.ValueFloat64(),
                Arch:             plan.Requirements.CPU.Arch.ValueString(),
                Flags:            cpuFlags,
                Slots:            int(plan.Requirements.CPU.Slots.ValueInt32()),
                AllowSMT:         plan.Requirements.CPU.AllowSMT.ValueBool(),
                FullPhysical:     plan.Requirements.CPU.FullPhysical.ValueBool(),
                MaxOverprovision: int(plan.Requirements.CPU.MaxOverprovision.ValueInt32()),
            },
            Memory: ComputeMemory{
                Capacity:   int(plan.Requirements.Memory.Capacity.ValueInt32()),
                RequireECC: plan.Requirements.Memory.RequireECC.ValueBool(),
            },
            PCI: ComputePCI{
                Devices: pciDevices,
            },
            Misc: ComputeMisc{
                OSFamily:  plan.Requirements.Misc.OSFamily.ValueString(),
                OSFlavour: plan.Requirements.Misc.OSFlavour.ValueString(),
            },
        },
        VMName: plan.VMName.ValueString(),
        Region: plan.Region.ValueString(),
        CloudInit: CloudInitConfig{
            Template:  plan.CloudInit.Template.ValueString(),
            Variables: cloudinitVariables,
        },
        AdditionalInfo: AdditionalInfoConfig{
            Service: plan.AdditionalInfo.Service.ValueString(),
        },
        BillingFrequency: plan.BillingFrequency.ValueString(),
        Volumes: volumes,
    }

    response, err := r.client.CreateComputeResource(ctx, computeResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating compute resource",
            "Could not create compute resource, unexpected error: " + err.Error(),
        )
        return
    }

    plan.BillingUUID = types.StringValue(response.BillingUUID)
    plan.PaymentURL = types.StringValue(response.PaymentURL)

    for volumeIndex, volume := range plan.Volumes {
        volume.Priority = types.Int32Null()
        volume.VolumeUUID = types.StringNull()
        volume.ServerURL = types.StringNull()
        volume.LastUpdated = types.StringNull()
        plan.Volumes[volumeIndex] = volume
    }

    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    // Set state to fully populated data
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *computeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *computeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *computeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
