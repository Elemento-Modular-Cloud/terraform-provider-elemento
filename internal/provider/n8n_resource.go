package provider

import (
	"context"
	"fmt"
	"time"

	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &n8nResource{}
	_ resource.ResourceWithConfigure = &n8nResource{}
)

func NewN8NResource() resource.Resource {
	return &n8nResource{}
}

type n8nResource struct{
	client *client.ElementoClient
}

type n8nResourceModel struct {
	VMName           types.String                `tfsdk:"vm_name"`
	CPUCores         types.Int32                 `tfsdk:"cpu_cores"`
	RAMSizeMB        types.Int32                 `tfsdk:"ram_size_mb"`
	Username         types.String                `tfsdk:"username"`
	Password         types.String                `tfsdk:"password"`
	SSHKey           types.String                `tfsdk:"ssh_key"`
    Region           types.String                `tfsdk:"region"`
    BillingFrequency types.String                `tfsdk:"billing_frequency"`
    BillingUUID      types.String                `tfsdk:"billing_uuid"`
    PaymentURL       types.String                `tfsdk:"payment_url"`
    LastUpdated      types.String                `tfsdk:"last_updated"`
    UUID             types.String                `tfsdk:"uuid"`
    State            types.String                `tfsdk:"state"`
    ServiceURL       types.String                `tfsdk:"service_url"`
    Status           types.String                `tfsdk:"status"`
}

func (r *n8nResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_n8n"
}

func (r *n8nResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
		    "vm_name": schema.StringAttribute{
				Required: true,
				Description: "The name of the virtual machine for the n8n resource.",
				PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
				},
			},
			"cpu_cores": schema.Int32Attribute{
                Required: true,
                Description: "The number of CPU cores for the n8n resource.",
                PlanModifiers: []planmodifier.Int32{
                    int32planmodifier.RequiresReplace(),
                },
            },
            "ram_size_mb": schema.Int32Attribute{
                Required: true,
                Description: "The amount of RAM in MB for the n8n resource.",
                PlanModifiers: []planmodifier.Int32{
                    int32planmodifier.RequiresReplace(),
                },
            },
            "username": schema.StringAttribute{
                Required: true,
                Description: "The username for the n8n resource.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "password": schema.StringAttribute{
                Required: true,
                Sensitive: true,
                Description: "The password for the n8n resource.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "ssh_key": schema.StringAttribute{
                Optional: true,
                Sensitive: true,
                Description: "The SSH key for the n8n resource.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "region": schema.StringAttribute{
                Required: true,
                Description: "The region where the n8n resource will be created.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "billing_frequency": schema.StringAttribute{
                Required: true,
                Description: "The billing frequency for the n8n resource (e.g., 'hour', 'month').",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "billing_uuid": schema.StringAttribute{
                Computed: true,
                Description: "The billing UUID for the n8n resource, assigned after creation.",

            },
            "payment_url": schema.StringAttribute{
                Computed: true,
                Description: "The payment URL for the n8n resource, assigned after creation.",
            },
            "last_updated": schema.StringAttribute{
                Computed: true,
                Description: "The timestamp of the last update to the n8n resource.",
            },
            "uuid": schema.StringAttribute{
                Computed: true,
                Description: "The UUID of the n8n resource, assigned after creation.",
            },
            "service_url": schema.StringAttribute{
                Computed: true,
                Description: "The service URL for the n8n resource, assigned after creation.",
            },
            "state": schema.StringAttribute{
                Computed: true,
                Description: "The state of the n8n resource (e.g., 'ERROR', 'SUCCESS').",
            },
            "status": schema.StringAttribute{
                Computed: true,
                Description: "The status of the n8n resource (e.g., 'paid', 'provisioning', 'running').",
            },
		},
	}
}

func (r *n8nResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *n8nResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan n8nResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    authentication := &client.ComputeAuthentication{
        Username: plan.Username.ValueString(),
        Password: plan.Password.ValueString(),
        SSHKey:   plan.SSHKey.ValueStringPointer(),
    }

    n8nResource := client.ComputeResource{
        Requirements: client.ComputeRequirements{
            CPU: client.ComputeCPU{
                SMTEnabled:       false,
                MinFrequency:     1,
                Arch:             []string{"X86_64"},
                Flags:            []string{},
                Slots:            int(plan.CPUCores.ValueInt32()),
                AllowSMT:         false,
                FullPhysical:     true,
                MaxOverprovision: 1,
            },
            Memory: client.ComputeMemory{
                Capacity:   int(plan.RAMSizeMB.ValueInt32()),
                RequireECC: false,
            },
            PCI: client.ComputePCI{
                Devices: map[string]int{},
            },
            Misc: client.ComputeMisc{
                OSFamily:  "linux",
                OSFlavour: "ubuntu",
            },
        },
        Authentication: authentication,
        VMName: plan.VMName.ValueString(),
        Region: plan.Region.ValueString(),
        CloudInit: client.CloudInitConfig{
            Template:  "n8n",
            Variables: map[string]string{},
        },
        AdditionalInfo: client.AdditionalInfoConfig{
            Service: "n8n",
        },
        BillingFrequency: plan.BillingFrequency.ValueString(),
        Volumes: []client.StorageResourceMin{
            {
                Name:      "n8n-main-disk",
                Size:      64,
                Private:   true,
                Shareable: false,
                Bootable:  false,
                Readonly:  false,
            },
        },
    }

    canCreateResponse, err := r.client.CanCreateComputeResource(n8nResource)
    ValidateCanCreate(canCreateResponse, &diags, "n8n", err)

    response, err := r.client.CreateComputeResource(n8nResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating n8n Resource",
            "Could not create n8n resource, unexpected error: " + err.Error(),
        )
        return
    }

    switch res := response.(type) {
        case *client.PublicCreateResponse:
            plan.BillingUUID = types.StringValue(res.BillingUUID)
            plan.PaymentURL  = types.StringValue(res.PaymentURL)
            plan.UUID        = types.StringNull()

            status, err := r.client.GetResourceStatus(res.BillingUUID)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Error Reading n8n Resource Status",
                    "Could not read n8n resource status: "+err.Error(),
                )
                return
            }

            plan.Status = types.StringValue(*status)
        case *client.PrivateCreateResponse:
            plan.BillingUUID = types.StringNull()
            plan.PaymentURL  = types.StringNull()
            plan.UUID        = types.StringValue(res.ServiceUUID)
            plan.Status      = types.StringValue(res.Status)
    }

    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    plan.State       = types.StringNull()
    plan.ServiceURL  = types.StringNull()

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *n8nResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state n8nResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    resources, err := r.client.ListN8NResources()
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading n8n Resources",
            "Could not list n8n resources: "+err.Error(),
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
        status, err = r.client.GetResourceStatus(state.BillingUUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading n8n Resource Status",
                "Could not read n8n resource status: "+err.Error(),
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

    state.Status      = types.StringValue(*status)
    state.VMName      = types.StringValue(resource.Requirements.VMName)
    state.CPUCores    = types.Int32Value(int32(resource.Requirements.Slots))
    state.RAMSizeMB   = types.Int32Value(int32(resource.Requirements.RamSize))
    state.VMName      = types.StringValue(resource.Requirements.VMName)
    state.Region      = types.StringValue(resource.Region)
    state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    state.UUID        = types.StringValue(resource.UUID)
    state.PaymentURL  = types.StringNull()

    credentials, err := r.client.GetN8NCredentials(resource.UUID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading n8n Credentials",
            "Could not read n8n credentials: "+err.Error(),
        )
        return
    }

    state.State      = types.StringValue(credentials.State)
    state.ServiceURL = types.StringValue(credentials.URL)

    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *n8nResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *n8nResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state n8nResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.client.DeleteComputeResource(state.UUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Deleting n8n Resource",
            "Could not delete n8n resource, unexpected error: " + err.Error(),
        )
        return
    }
}
