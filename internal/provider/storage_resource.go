package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
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
}

func (r *storageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func volumeAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "creator_uuid": schema.StringAttribute{
            Optional: true,
            Computed: true,
        },
        "name": schema.StringAttribute{
            Required: true,
        },
        "size": schema.Int64Attribute{
            Required: true,
        },
        "private": schema.BoolAttribute{
	       	Required: true,
        },
        "shareable": schema.BoolAttribute{
	       	Required: true,
        },
        "bootable": schema.BoolAttribute{
	       	Required: true,
        },
        "clonable": schema.BoolAttribute{
           	Optional: true,
            Computed: true,
        },
        "alg": schema.StringAttribute{
           	Optional: true,
        },
        "format": schema.StringAttribute{
           	Optional: true,
        },
        "bus": schema.StringAttribute{
           	Optional: true,
        },
        "fs": schema.StringAttribute{
           	Optional: true,
        },
        "hardware_device": schema.StringAttribute{
           	Optional: true,
        },
        "tags": schema.ListAttribute{
            ElementType: types.StringType,
            Optional: true,
            Computed: true,
            Default: listdefault.StaticValue(
				types.ListValueMust(
					types.StringType,
					[]attr.Value{},
				),
			),
        },
        "region": schema.StringAttribute{
           	Optional: true,
            Computed: true,
        },
        "readonly": schema.BoolAttribute{
	       	Required: true,
        },
        "priority": schema.Int32Attribute{
            Optional: true,
	       	Computed: true,
        },
        "volume_uuid": schema.StringAttribute{
 	       	Computed: true,
        },
        "last_updated": schema.StringAttribute{
        	Computed: true,
        },
        "billing_frequency": schema.StringAttribute{
            Optional: true,
        },
        "billing_uuid": schema.StringAttribute{
            Computed: true,
        },
        "payment_url": schema.StringAttribute{
            Computed: true,
        },
    }
}

func (r *storageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    attributes := volumeAttributes()
    attributes["region"] = schema.StringAttribute{
        Required: true,
    }
    attributes["billing_frequency"] = schema.StringAttribute{
        Required: true,
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

    // TODO: Add cancreate handling

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

    if resource == nil {
        resp.State.RemoveResource(ctx)
        return
    }

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
