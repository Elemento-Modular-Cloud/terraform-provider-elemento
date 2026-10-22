package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &storageResource{}
	_ resource.ResourceWithConfigure = &storageResource{}
)

func NewStorageResource() resource.Resource {
	return &storageResource{}
}

type storageResource struct{
	client *client.ElementoClient
}

type storageResourceModel struct {
    CreatorUUID      types.String   `tfsdk:"creator_uuid"`
    Name             types.String   `tfsdk:"name"`
    Size             types.Int64    `tfsdk:"size"`
    Private          types.Bool     `tfsdk:"private"`
    Shareable        types.Bool     `tfsdk:"shareable"`
    Bootable         types.Bool     `tfsdk:"bootable"`
    Clonable         types.Bool     `tfsdk:"clonable"`
    Algorithm        types.String   `tfsdk:"alg"`
    Format           types.String   `tfsdk:"format"`
    Bus              types.String   `tfsdk:"bus"`
    FS               types.String   `tfsdk:"fs"`
    HardwareDevice   types.String   `tfsdk:"hardware_device"`
    Tags             types.List     `tfsdk:"tags"`
    Region           types.String   `tfsdk:"region"`
    Readonly         types.Bool     `tfsdk:"readonly"`
    Priority         types.Int32    `tfsdk:"priority"`
    UUID             types.String   `tfsdk:"volume_uuid"`
    LastUpdated      types.String   `tfsdk:"last_updated"`
    BillingFrequency types.String   `tfsdk:"billing_frequency"`
    BillingUUID      types.String   `tfsdk:"billing_uuid"`
    PaymentURL       types.String   `tfsdk:"payment_url"`
    Status           types.String   `tfsdk:"status"`
}

func (r *storageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func volumeAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "creator_uuid": schema.StringAttribute{
            Optional: true,
            Computed: true,
            Description: "The UUID of the creator of the storage resource. If not provided, it will be set to the current user's UUID.",
        },
        "name": schema.StringAttribute{
            Required: true,
            Description: "The name of the storage resource.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "size": schema.Int64Attribute{
            Required: true,
            Description: "The size of the storage resource in bytes.",
            PlanModifiers: []planmodifier.Int64 {
                int64planmodifier.RequiresReplace(),
            },
        },
        "private": schema.BoolAttribute{
	       	Required: true,
            Description: "Whether the storage resource is private.",
            PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "shareable": schema.BoolAttribute{
	       	Required: true,
			Description: "Whether the storage resource is shareable.",
			PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "bootable": schema.BoolAttribute{
	       	Required: true,
			Description: "Whether the storage resource is bootable.",
			PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "clonable": schema.BoolAttribute{
           	Optional: true,
            Computed: true,
            Description: "Whether the storage resource is clonable.",
            PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "alg": schema.StringAttribute{
           	Optional: true,
            Description: "The algorithm used for the storage resource clone.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
       },
        "format": schema.StringAttribute{
           	Optional: true,
            Description: "The format of the storage resource.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "bus": schema.StringAttribute{
           	Optional: true,
            Description: "The bus type of the storage resource.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "fs": schema.StringAttribute{
           	Optional: true,
            Description: "The underlying filesystem type of the storage resource.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "hardware_device": schema.StringAttribute{
           	Optional: true,
            Description: "Tne required underlying hardware device for the storage resource.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "tags": schema.ListAttribute{
            ElementType: types.StringType,
            Description: "A list of tags associated with the storage resource.",
            Optional: true,
            Computed: true,
            Default: listdefault.StaticValue(
				types.ListValueMust(
					types.StringType,
					[]attr.Value{},
				),
			),
            PlanModifiers: []planmodifier.List{
                listplanmodifier.RequiresReplace(),
            },
        },
        "region": schema.StringAttribute{
           	Optional: true,
            Computed: true,
            Description: "The region where the storage resource is located.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "readonly": schema.BoolAttribute{
	       	Required: true,
            Description: "Whether the storage resource is read-only.",
            PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "priority": schema.Int32Attribute{
            Optional: true,
	       	Computed: true,
            Description: "The boot priority of the storage resource.",
            PlanModifiers: []planmodifier.Int32 {
                int32planmodifier.RequiresReplace(),
            },
        },
        "volume_uuid": schema.StringAttribute{
 	       	Computed: true,
            Description: "The UUID of the storage resource.",
        },
        "last_updated": schema.StringAttribute{
        	Computed: true,
            Description: "The last time the storage resource was updated.",
        },
        "billing_frequency": schema.StringAttribute{
            Optional: true,
            Description: "The frequency at which the compute resource should be billed. Valid values are 'day', 'week', 'month', and 'year'.",
        },
        "billing_uuid": schema.StringAttribute{
            Computed: true,
            Description: "The billing UUID of the storage resource, assigned by the Elemento API after creation.",
        },
        "payment_url": schema.StringAttribute{
            Computed: true,
            Description: "The payment URL for the storage resource, assigned by the Elemento API after creation.",
        },
        "status": schema.StringAttribute{
            Computed: true,
            Description: "The current status of the storage resource, computed by the Elemento API.",
        },
    }
}

func (r *storageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    attributes := volumeAttributes()
    attributes["region"] = schema.StringAttribute{
        Required: true,
        PlanModifiers: []planmodifier.String {
           stringplanmodifier.RequiresReplace(),
        },
    }
    attributes["billing_frequency"] = schema.StringAttribute{
        Required: true,
        Description: "The frequency at which the compute resource should be billed. Valid values are 'day', 'week', 'month', and 'year'.",
        PlanModifiers: []planmodifier.String {
           stringplanmodifier.RequiresReplace(),
        },
    }

    resp.Schema = schema.Schema{
		Attributes: volumeAttributes(),
	}
}

func (r *storageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan storageResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var tags []string
    diags = plan.Tags.ElementsAs(ctx, &tags, false)
    resp.Diagnostics.Append(diags...)

    var priority *int
    if !plan.Priority.IsNull() {
        priorityValue := int(plan.Priority.ValueInt32())
        priority = &priorityValue
    }

    if resp.Diagnostics.HasError() {
        return
    }

    storageResource := client.StorageResource{
        CreatorUUID:      plan.CreatorUUID.ValueString(),
        Name:             plan.Name.ValueString(),
        Size:             int(plan.Size.ValueInt64()),
        Private:          plan.Private.ValueBool(),
        Shareable:        plan.Shareable.ValueBool(),
        Bootable:         plan.Bootable.ValueBool(),
        Clonable:         plan.Clonable.ValueBoolPointer(),
        Algorithm:        plan.Algorithm.ValueStringPointer(),
        Format:           plan.Format.ValueStringPointer(),
        Bus:              plan.Bus.ValueStringPointer(),
        FS:               plan.FS.ValueStringPointer(),
        HardwareDevice:   plan.HardwareDevice.ValueStringPointer(),
        Tags:             tags,
        Region:           plan.Region.ValueString(),
        Readonly:         plan.Readonly.ValueBool(),
        Priority:         priority,
        BillingFrequency: plan.BillingFrequency.ValueString(),
    }


    canCreateRequest := client.CanCreateStorageRequest{
        Size:             int(plan.Size.ValueInt64()),
        Region:           plan.Region.ValueString(),
        BillingFrequency: plan.BillingFrequency.ValueString(),
        Benchmark:        client.CanCreateStorageBenchmark{
            ReadIOPS:         0,
            WriteIOPS:        0,
            ReadMBBandwidth:  0,
            WriteMBBandwidth: 0,
        },
        HardwareDevice:   plan.HardwareDevice.ValueStringPointer(),
        Filesystem:       plan.FS.ValueStringPointer(),
        Ceph:             nil,
    }

    canCreateResponse, err := r.client.CanCreateStorageResource(canCreateRequest)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Checking Storage Resource Creation",
            "Could not check if storage resource can be created, unexpected error: " + err.Error(),
        )
        return
    }

    if !canCreateResponse.CanCreate {
        for field, reason := range canCreateResponse.Reasons {
            resp.Diagnostics.AddError(
                "Storage Resource Creation Not Allowed",
                fmt.Sprintf("Cannot create storage resource due to field '%s' value: %s", field, reason),
            )
        }
        return
    }

    response, err := r.client.CreateStorageResource(storageResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating Storage Resource",
            "Could not create storage resource, unexpected error: " + err.Error(),
        )
        return
    }

    plan.BillingUUID = types.StringValue(response.BillingUUID)
    plan.PaymentURL = types.StringValue(response.PaymentURL)
    plan.UUID = types.StringNull()
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    plan.CreatorUUID = types.StringNull()

    status, err := r.client.GetResourceStatus(response.BillingUUID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading Storage Resource Status",
            "Could not read storage resource status: "+err.Error(),
        )
        return
    }

    plan.Status = types.StringValue(*status)

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *storageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state storageResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var resource *client.StorageResourceResponse
    if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
        resources, err := r.client.ListStorageResources(state.CreatorUUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading Storage Resources",
                "Could not list storage resources: "+err.Error(),
            )
            return
        }

        for _, res := range resources {
            if res.BillingUUID == state.BillingUUID.ValueString() {
                resource = &res
                break
            }
        }
    } else {
        res, err := r.client.GetStorageResource(state.CreatorUUID.ValueString(), state.UUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading Storage Resource",
                "Could not read storage resource: "+err.Error(),
            )
            return
        }
        resource = res
    }

    status, err := r.client.GetResourceStatus(state.BillingUUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading Storage Resource Status",
            "Could not read storage resource status: "+err.Error(),
        )
        return
    }

    if resource == nil && (*status == "deleted" || *status == "error" || *status == "canceled") {
        resp.State.RemoveResource(ctx)
        return
    }

    state.Status      = types.StringValue(*status)
    state.Name        = types.StringValue(resource.Name)
    state.Private     = types.BoolValue(resource.Private)
    state.Shareable   = types.BoolValue(resource.Shareable)
    state.Bootable    = types.BoolValue(resource.Bootable)
    state.Readonly    = types.BoolValue(resource.Readonly)
    state.Clonable    = types.BoolValue(resource.Clonable)
    state.Size        = types.Int64Value(int64(resource.Size))
    state.Priority    = types.Int32Value(int32(resource.Priority))
    state.Tags, diags = types.ListValueFrom(ctx, types.StringType, resource.Tags)
    resp.Diagnostics.Append(diags...)

    state.Region      = types.StringValue(resource.Region)
    state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    state.UUID        = types.StringValue(resource.UUID)
    state.PaymentURL  = types.StringNull()

    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

}

func (r *storageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *storageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state storageResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.client.DeleteStorageResource(state.CreatorUUID.ValueString(), state.UUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddWarning(
            "Error Deleting Storage Resource",
            "Could not delete storage resource, unexpected error: " + err.Error(),
        )
        return
    }
}
