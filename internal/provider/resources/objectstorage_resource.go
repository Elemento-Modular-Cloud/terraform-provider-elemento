package resources

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &objectStorageResource{}
	_ resource.ResourceWithConfigure   = &objectStorageResource{}
	_ resource.ResourceWithImportState = &objectStorageResource{}
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

func (r *objectStorageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema(ctx)

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "Name of the bucket.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["size_tb"] = schema.Int64Attribute{
		Required:    true,
		Description: "Bucket size in TB.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["active_storage"] = schema.Int32Attribute{
		Computed:    true,
		Description: "Storage currently in use.",
	}
	attributes["endpoint"] = schema.StringAttribute{
		Computed:    true,
		Description: "Endpoint URL of the bucket.",
	}
	attributes["account_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Storage account name.",
	}
	attributes["container_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Container name.",
	}
	attributes["sas_token"] = schema.StringAttribute{
		Computed:    true,
		Description: "SAS token granting access to the bucket.",
		Sensitive:   true,
	}
	attributes["azure_cli_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "Azure CLI command that accesses the bucket.",
		Sensitive:   true,
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

	createTimeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	objectStorageResource := client.ObjectStorageRequest{
		Name:             plan.Name.ValueString(),
		SizeTB:           int(plan.SizeTB.ValueInt64()),
		Region:           plan.Region.ValueString(),
		BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
	}

	canCreateResponse, err := r.Client.CanCreateObjectStorageResource(ctx, objectStorageResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "ObjectStorage", err) {
		return
	}

	response, err := r.Client.CreateObjectStorageResource(ctx, objectStorageResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating ObjectStorage Resource",
			"Could not create objectstorage resource, unexpected error: "+err.Error(),
		)
		return
	}

	mustRead := false
	switch res := response.(type) {
	case client.PublicCreateResponse:
		plan.BillingUUID = types.StringValue(res.BillingUUID)
		plan.PaymentURL = types.StringValue(res.PaymentURL)
		plan.UUID = types.StringNull()

		status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Reading ObjectStorage Resource Status",
				"Could not read objectstorage resource status: "+err.Error(),
			)
			status = nil
		}

		plan.Status = types.StringPointerValue(status)
	case client.PrivateCreateResponse:
		plan.BillingUUID = types.StringNull()
		plan.PaymentURL = types.StringNull()
		plan.UUID = types.StringValue(res.ServiceUUID)
		plan.Status = types.StringValue(res.Status)

		mustRead = true
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

	if mustRead {
    	stateConf := common.GetCreateStateConf(ctx, r.Client, createTimeout, &plan)

		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			resp.Diagnostics.AddError("Timeout waiting for resource creation", err.Error())
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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

	deleteTimeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusCode, err := r.Client.DeleteObjectStorageResource(ctx, state.UUID.ValueString())
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

	stateConf := common.GetDeleteStateConf(ctx, r.Client, deleteTimeout, &state)
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Resource Deletion",
			fmt.Sprintf("Could not confirm deletion of resource %s: %s", state.UUID.ValueString(), err.Error()),
		)
		return
	}
}

func (r *objectStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
