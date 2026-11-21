package resources

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var (
	_ resource.Resource = &storageResource{}
	_ resource.ResourceWithConfigure = &storageResource{}
	_ resource.ResourceWithImportState = &storageResource{}
)

func NewStorageResource() resource.Resource {
	return &storageResource{}
}

type storageResource struct{
	common.BaseResource
}

func (r *storageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func volumeAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "creator_uuid": schema.StringAttribute{
            Computed: true,
            Description: "UUID of the account that created the resource.",
        },
        "name": schema.StringAttribute{
            Required: true,
            Description: "Volume name.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "size": schema.Int64Attribute{
            Required: true,
            Description: "Volume size in GB.",
            PlanModifiers: []planmodifier.Int64 {
                int64planmodifier.RequiresReplace(),
            },
        },
        "private": schema.BoolAttribute{
	       	Required: true,
            Description: "Whether the volume is private to its owner.",
            PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "shareable": schema.BoolAttribute{
	       	Required: true,
			Description: "Whether the volume can be attached to more than one VM.",
			PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "bootable": schema.BoolAttribute{
	       	Required: true,
			Description: "Whether the volume can be used as a boot disk.",
			PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "clonable": schema.BoolAttribute{
           	Optional: true,
            Computed: true,
            Description: "Whether the volume can be cloned.",
            PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
            Default: booldefault.StaticBool(true),
        },
        "alg": schema.StringAttribute{
           	Optional: true,
            Description: "Algorithm used to clone the volume.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
       },
        "format": schema.StringAttribute{
           	Optional: true,
            Description: "Disk image format.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "bus": schema.StringAttribute{
           	Optional: true,
            Description: "Bus type the volume is attached with.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "fs": schema.StringAttribute{
           	Optional: true,
            Description: "Filesystem type of the volume.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "hardware_device": schema.StringAttribute{
           	Optional: true,
            Description: "Hardware device the volume requires.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "tags": schema.ListAttribute{
            ElementType: types.StringType,
            Description: "Tags to attach to the volume.",
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
            Description: "Region to deploy in. Changing it forces a new resource.",
            PlanModifiers: []planmodifier.String {
                stringplanmodifier.RequiresReplace(),
            },
        },
        "readonly": schema.BoolAttribute{
	       	Required: true,
            Description: "Whether the volume is attached read-only.",
            PlanModifiers: []planmodifier.Bool {
                boolplanmodifier.RequiresReplace(),
            },
        },
        "priority": schema.Int32Attribute{
            Optional: true,
	       	Computed: true,
            Description: "Boot priority.",
            PlanModifiers: []planmodifier.Int32 {
                int32planmodifier.RequiresReplace(),
            },
        },
        "uuid": schema.StringAttribute{
 	       	Computed: true,
            Description: "Service UUID.",
        },
        "last_updated": schema.StringAttribute{
        	Computed: true,
            Description: "Timestamp of the last update.",
        },
        "billing_frequency": schema.StringAttribute{
            Optional: true,
            Description: "Billing period, e.g. `day`, `month` or `year`. Changing it forces a new resource.",
        },
        "billing_uuid": schema.StringAttribute{
            Computed: true,
            Description: "Portal billing UUID.",
        },
        "payment_url": schema.StringAttribute{
            Computed: true,
            Description: "URL where the payment for the resource is completed.",
        },
        "status": schema.StringAttribute{
            Computed: true,
            Description: "Current status of the resource.",
        },
		"creation_date": schema.StringAttribute{
			Computed:    true,
			Description: "Creation timestamp.",
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
        Description: "Billing period, e.g. `day`, `month` or `year`. Changing it forces a new resource.",
        PlanModifiers: []planmodifier.String {
           stringplanmodifier.RequiresReplace(),
        },
    }

    resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *storageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan common.StorageResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

	createTimeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

    var tags []string
    diags = plan.Tags.ElementsAs(ctx, &tags, false)
    resp.Diagnostics.Append(diags...)

    if plan.Priority.IsUnknown() {
    	plan.Priority = types.Int32Null()
    }

    var priority *int
    if !plan.Priority.IsNull() {
        priorityValue := int(plan.Priority.ValueInt32())
        priority = &priorityValue
    }

    if resp.Diagnostics.HasError() {
        return
    }

    storageResource := client.StorageResource{
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

    canCreateResponse, err := r.Client.CanCreateStorageResource(ctx, canCreateRequest)
    if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "Storage", err) {
	    return
    }

    response, err := r.Client.CreateStorageResource(ctx, &storageResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating Storage Resource",
            "Could not create storage resource, unexpected error: " + err.Error(),
        )
        return
    }

    mustRead := false
    switch res := response.(type) {
        case client.PublicCreateResponse:
            plan.BillingUUID = types.StringValue(res.BillingUUID)
            plan.PaymentURL  = types.StringValue(res.PaymentURL)
            plan.UUID        = types.StringNull()

            status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
            if err != nil {
                resp.Diagnostics.AddWarning(
                    "Error Reading Storage Resource Status",
                    "Could not read storage resource status: "+err.Error(),
                )
                status = nil
            }

            plan.Status = types.StringPointerValue(status)
        case client.PrivateCreateResponse:
            plan.BillingUUID = types.StringNull()
            plan.PaymentURL  = types.StringNull()
            plan.UUID        = types.StringValue(res.ServiceUUID)
            plan.Status      = types.StringValue(res.Status)

            mustRead = true
    }

    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    plan.CreatorUUID = types.StringNull()
    plan.CreationDate = types.StringNull()

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    if mustRead {
    	stateConf := &retry.StateChangeConf{
     		Pending: []string{"CREATING"},
       		Target: []string{"READY"},
         	Delay: 5 * time.Second,
          	MinTimeout: 3 * time.Second,
           	Timeout: createTimeout,
            Refresh: func() (any, string, error) {
            	found, diags := plan.Read(ctx, r.Client, func(_ context.Context, _ bool) {})
             	if diags.HasError() {
             		return nil, "", fmt.Errorf("%s: %s", diags.Errors()[0].Summary(), diags.Errors()[0].Detail())
             	}
              	if !found {
               		return &plan, "CREATING", nil
                }
                return &plan, "READY", nil
            },
     	}

      	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
       		resp.Diagnostics.AddError("Timeout waiting for resource creation", err.Error())
         	return
       	}

        resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
    }
}

func (r *storageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var model common.StorageResourceModel
    diags := req.State.Get(ctx, &model)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    found, diags := model.Read(ctx, r.Client, common.RemoveResource(resp))
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() || !found {
    	return
    }
    model.PaymentURL = types.StringNull()

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

func (r *storageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *storageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state common.StorageResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    if state.UUID.IsNull() {
	    resp.Diagnostics.AddError(
		    "Error Deleting Storage Resource",
		    "Could not delete storage resource: resource uuid is null.",
	    )
		return
    }

    statusCode, err := r.Client.DeleteStorageResource(ctx, state.UUID.ValueString())
    if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting Storage Resource",
				"Could not delete storage resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Storage Resource",
				"Could not delete storage resource, unexpected error: " + err.Error(),
			)
		}
		return
    }
}

func (r *storageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
