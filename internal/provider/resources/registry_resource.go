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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &registryResource{}
	_ resource.ResourceWithConfigure   = &registryResource{}
	_ resource.ResourceWithImportState = &registryResource{}
)

func NewRegistryResource() resource.Resource {
	return &registryResource{}
}

type registryResource struct {
	common.BaseResource
}

func (r *registryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (r *registryResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema(ctx)

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "Name of the registry.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Required:    true,
		Description: "Storage size in GB.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["is_public"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether the registry is reachable from the public internet.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["repository_uri"] = schema.StringAttribute{
		Computed:    true,
		Description: "URI of the repository.",
	}
	attributes["registry_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the registry.",
	}
	attributes["image_count"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Number of images stored.",
	}
	attributes["used_bytes"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Storage in use, in bytes.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "Access key ID for the registry.",
		Sensitive:   true,
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "Secret access key for the resource.",
		Sensitive:   true,
	}
	attributes["registry_type"] = schema.StringAttribute{
		Computed:    true,
		Description: "Registry type.",
	}
	attributes["login_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "Command that logs in to the registry.",
		Sensitive:   true,
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *registryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.RegistryResourceModel
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

	registryResource := client.RegistryRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:     plan.Name.ValueString(),
		SizeGB:   plan.SizeGB.ValueInt64(),
		IsPublic: plan.IsPublic.ValueBool(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(ctx, registryResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "Registry", err) {
		return
	}

	response, err := r.Client.CreateResource(ctx, registryResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Registry Resource",
			"Could not create registry resource, unexpected error: "+err.Error(),
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
				"Error Reading Registry Resource Status",
				"Could not read registry resource status: "+err.Error(),
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
	default:
		resp.Diagnostics.AddError(
			"Unexpected API Response Type",
			fmt.Sprintf("CreateRegistryResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.RepositoryURI = types.StringNull()
	plan.RegistryID = types.StringNull()
	plan.ImageCount = types.Int64Null()
	plan.UsedBytes = types.Int64Null()
	plan.AccessKeyID = types.StringNull()
	plan.SecretAccessKey = types.StringNull()
	plan.RegistryType = types.StringNull()
	plan.LoginCommand = types.StringNull()

	plan.State = types.StringNull()

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

func (r *registryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.RegistryResourceModel
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

func (r *registryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *registryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.RegistryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error Deleting Registry Resource",
			"Could not delete registry resource: resource uuid is null.",
		)
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusCode, err := r.Client.DeleteResource(ctx, state.UUID.ValueString(), "registry")
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting Registry Resource",
				"Could not delete registry resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Registry Resource",
				"Could not delete registry resource, unexpected error: "+err.Error(),
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

func (r *registryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
