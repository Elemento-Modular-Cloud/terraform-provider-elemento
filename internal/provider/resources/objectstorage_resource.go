package resources

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &objectStorageResource{}
	_ resource.ResourceWithConfigure = &objectStorageResource{}
)

func NewObjectStorageResource() resource.Resource {
	return &objectStorageResource{}
}

type objectStorageResource struct {
	common.BaseResource
}

func (r *objectStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_objectstorage"
}

func (r *objectStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the object storage resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["size_tb"] = schema.Int64Attribute{
		Required:    true,
		Description: "The size of the object storage resource in TB.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},

	}
	attributes["active_storage"] = schema.Int32Attribute{
		Computed:    true,
		Description: "The amount of active storage in the object storage resource.",
	}
	attributes["endpoint"] = schema.StringAttribute{
		Computed:    true,
		Description: "The endpoint URL for the object storage resource.",
	}
	attributes["account_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the storage account.",
	}
	attributes["container_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the storage container.",
	}
	attributes["sas_token"] = schema.StringAttribute{
		Computed:    true,
		Description: "The SAS token for accessing the storage account.",
		Sensitive: true,
	}
	attributes["azure_cli_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "The Azure CLI command to access the storage account.",
		Sensitive: true,
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *objectStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.ObjectStorageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	objectStorageResource := client.ObjectStorageRequest{
		Name:             plan.Name.ValueString(),
		SizeTB:           int(plan.SizeTB.ValueInt64()),
		Region:           plan.Region.ValueString(),
		BillingFrequency: plan.BillingFrequency.ValueString(),
	}

	canCreateResponse, err := r.Client.CanCreateObjectStorageResource(objectStorageResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "ObjectStorage", err) {
		return
	}

	response, err := r.Client.CreateObjectStorageResource(objectStorageResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating ObjectStorage Resource",
			"Could not create objectstorage resource, unexpected error: "+err.Error(),
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

		status, err := r.Client.GetResourceStatus(res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Reading ObjectStorage Resource Status",
				"Could not read objectstorage resource status: "+err.Error(),
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

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.ActiveStorage = types.Int32Null()
	plan.Endpoint = types.StringNull()
	plan.AccountName = types.StringNull()
	plan.ContainerName = types.StringNull()
	plan.SasToken = types.StringNull()
	plan.AzureCliCommand = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *objectStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.ObjectStorageResourceModel
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

func (r *objectStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *objectStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.ObjectStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error Deleting ObjectStorage Resource",
			"Could not delete objectstorage resource: resource uuid is null.",
		)
		return
	}

	statusCode, err := r.Client.DeleteObjectStorageResource(state.UUID.ValueString())
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting ObjectStorage Resource",
				"Could not delete objectstorage resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting ObjectStorage Resource",
				"Could not delete objectstorage resource, unexpected error: "+err.Error(),
			)
		}
		return
	}
}
