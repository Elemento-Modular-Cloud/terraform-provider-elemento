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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &aiserverlessResource{}
	_ resource.ResourceWithConfigure = &aiserverlessResource{}
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

func (r *aiserverlessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["model_name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the model to use for the aiserverless resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["enable_streaming"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether streaming should be enabled for the resource or not.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["model_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the model used by the resource.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ARN of the resource.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The access key ID for the resource.",
		Sensitive: true,
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "The secret access key of the resource.",
		Sensitive: true,
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

	aiserverlessResource := client.AIServerlessRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:            plan.Name.ValueString(),
		ModelName:       plan.ModelName.ValueString(),
		EnableStreaming: plan.EnableStreaming.ValueBool(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(aiserverlessResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "AIServerless", err) {
		return
	}

	response, err := r.Client.CreateResource(aiserverlessResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating AIServerless Resource",
			"Could not create aiserverless resource, unexpected error: "+err.Error(),
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
				"Error Reading AIServerless Resource Status",
				"Could not read aiserverless resource status: "+err.Error(),
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

	statusCode, err := r.Client.DeleteResource(state.UUID.ValueString(), "aiserverless")
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
