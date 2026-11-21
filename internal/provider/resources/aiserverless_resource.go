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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var (
	_ resource.Resource                = &aiserverlessResource{}
	_ resource.ResourceWithConfigure   = &aiserverlessResource{}
	_ resource.ResourceWithImportState = &aiserverlessResource{}
)

func NewAIServerlessResource() resource.Resource {
	return &aiserverlessResource{}
}

type aiserverlessResource struct {
	common.BaseResource
}

func (r *aiserverlessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aiserverless"
}

func (r *aiserverlessResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema(ctx)

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "Name of the AI serverless endpoint.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["model_name"] = schema.StringAttribute{
		Required:    true,
		Description: "Name of the model to serve.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["enable_streaming"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether responses are streamed.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["model_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the model being served.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "ARN of the resource.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "Access key ID for the resource.",
		Sensitive:   true,
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "Secret access key for the resource.",
		Sensitive:   true,
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *aiserverlessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.AIServerlessResourceModel
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

	aiserverlessResource := client.AIServerlessRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:            plan.Name.ValueString(),
		ModelName:       plan.ModelName.ValueString(),
		EnableStreaming: plan.EnableStreaming.ValueBool(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(ctx, aiserverlessResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "AIServerless", err) {
		return
	}

	response, err := r.Client.CreateResource(ctx, aiserverlessResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating AIServerless Resource",
			"Could not create aiserverless resource, unexpected error: "+err.Error(),
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
				"Error Reading AIServerless Resource Status",
				"Could not read aiserverless resource status: "+err.Error(),
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
			fmt.Sprintf("CreateAIServerlessResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.ModelID = types.StringNull()
	plan.ARN = types.StringNull()
	plan.AccessKeyID = types.StringNull()
	plan.SecretAccessKey = types.StringNull()

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

func (r *aiserverlessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.AIServerlessResourceModel
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

func (r *aiserverlessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *aiserverlessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.AIServerlessResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error Deleting AIServerless Resource",
			"Could not delete aiserverless resource: resource uuid is null.",
		)
		return
	}

	statusCode, err := r.Client.DeleteResource(ctx, state.UUID.ValueString(), "aiserverless")
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting AIServerless Resource",
				"Could not delete aiserverless resource: resource not found.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting AIServerless Resource",
				"Could not delete aiserverless resource, unexpected error: "+err.Error(),
			)
		}

		return
	}
}

func (r *aiserverlessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
