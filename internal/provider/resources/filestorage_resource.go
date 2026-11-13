package resources

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &fileStorageResource{}
	_ resource.ResourceWithConfigure = &fileStorageResource{}
)

func NewFileStorageResource() resource.Resource {
	return &fileStorageResource{}
}

type fileStorageResource struct {
	common.BaseResource
}

func (r *fileStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filestorage"
}

func (r *fileStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Required:    true,
		Description: "The size (in GB) of the resource.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["is_public"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether this resource is public or not.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["used_gb"] = schema.Float64Attribute{
		Computed:    true,
		Description: "The size (in GB) currently used by the resource.",
	}
	attributes["dns_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the DNS associated with the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}
	attributes["nfs_host"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs host associated with the resource",
	}
	attributes["nfs_path"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs path associated with the resource",
	}
	attributes["nfs_version"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs version associated with the resource",
	}
	attributes["mount_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "The command used to connect to the resource",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *fileStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.FileStorageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileStorageResource := client.FileStorageRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:   plan.Name.ValueString(),
		SizeGB: int64(plan.SizeGB.ValueInt64()),
		IsPublic: plan.IsPublic.ValueBool(),
	}

	canCreateResponse, err := r.Client.CanCreateFileStorageResource(ctx, fileStorageResource)
    if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "FileStorage", err) {
	    return
    }

	response, err := r.Client.CreateFileStorageResource(ctx, fileStorageResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating FileStorage Resource",
			"Could not create filestorage resource, unexpected error: "+err.Error(),
		)
		return
	}

	switch res := response.(type) {
	case client.PublicCreateResponse:
		plan.BillingUUID = types.StringValue(res.BillingUUID)
		plan.PaymentURL = types.StringValue(res.PaymentURL)
		plan.UUID = types.StringNull()

		diags = resp.State.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)

		status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Reading FileStorage Resource Status",
				"Could not read filestorage resource status: "+err.Error(),
			)
			return
		}

		plan.Status = types.StringValue(*status)
	case client.PrivateCreateResponse:
		plan.BillingUUID = types.StringNull()
		plan.PaymentURL = types.StringNull()
		plan.UUID = types.StringValue(res.ServiceUUID)
		plan.Status = types.StringValue(res.Status)
	default:
		resp.Diagnostics.AddError(
			"Unexpected API Response Type",
			fmt.Sprintf("CreateFileStorageResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.UsedGB = types.Float64Null()
	plan.DNSName = types.StringNull()
	plan.State = types.StringNull()
	plan.NFSHost = types.StringNull()
	plan.NFSPath = types.StringNull()
	plan.NFSVersion = types.StringNull()
	plan.MountCommand = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *fileStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.FileStorageResourceModel
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

func (r *fileStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *fileStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.FileStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error Deleting FileStorage Resource",
			"Could not delete filestorage resource: resource uuid is null.",
		)
		return
	}

	statusCode, err := r.Client.DeleteFileStorageResource(ctx, state.UUID.ValueString())
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting FileStorage Resource",
				"Could not delete filestorage resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting FileStorage Resource",
				"Could not delete filestorage resource, unexpected error: "+err.Error(),
			)
		}
		return
	}
}
