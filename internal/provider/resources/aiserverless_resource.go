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

type aiserverlessResourceModel struct {
	common.BaseResourceModel
	Name            types.String `tfsdk:"name"`
	ModelName       types.String `tfsdk:"model_name"`
	EnableStreaming types.Bool   `tfsdk:"enable_streaming"`
	ModelID         types.String `tfsdk:"model_id"`
	ARN             types.String `tfsdk:"arn"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
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
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "The secret access key of the resource.",
	}


	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *aiserverlessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aiserverlessResourceModel
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
		Name:        plan.Name.ValueString(),
		ModelName: plan.ModelName.ValueString(),
		EnableStreaming: plan.EnableStreaming.ValueBool(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(aiserverlessResource)
	common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "AIServerless", err)

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

		status, err := r.Client.GetResourceStatus(res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddError(
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
	var state aiserverlessResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resource *client.AIServerlessResponse
	if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.AIServerlessResponse](r.Client, "aiserverless")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading AIServerless Resources",
				"Could not list aiserverless resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.AIServerlessResponse](r.Client, state.UUID.ValueString(), "aiserverless")
		if err != nil && code != 404 {
			resp.Diagnostics.AddError(
				"Error Reading AIServerless Resource"+string(code),
				"Could not read aiserverless resource: "+err.Error(),
			)
			return
		}
		resource = res
	}

	billing_uuid := state.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		var err error
		status, err = r.Client.GetResourceStatus(state.BillingUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Compute Resource Status",
				"Could not read compute resource status: "+err.Error(),
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

	state.Status = types.StringPointerValue(status)
	state.Name = types.StringValue(resource.Name)
	state.Region = types.StringValue(resource.Region)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	state.UUID = types.StringValue(resource.ServiceUUID)
	state.CreatorUUID = types.StringValue(resource.CreatorUUID)
	state.CreationDate = types.StringValue(resource.CreationDate)
	state.ModelName = types.StringValue(resource.ModelName)
	state.EnableStreaming = types.BoolValue(resource.EnableStreaming)
	state.ModelID = types.StringValue(resource.ModelID)
	state.ARN = types.StringValue(resource.ARN)
	state.PaymentURL = types.StringNull()

	credentials, err := r.Client.GetAIServerlessCredentials(resource.ServiceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AIServerless Credentials",
			"Could not read aiserverless credentials: "+err.Error(),
		)
		return
	}

	state.AccessKeyID = types.StringValue(credentials.AccessKeyID)
	state.SecretAccessKey = types.StringValue(credentials.SecretAccessKey)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *aiserverlessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *aiserverlessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state aiserverlessResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.DeleteResource(state.UUID.ValueString(), "aiserverless")
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting AIServerless Resource",
			"Could not delete aiserverless resource, unexpected error: "+err.Error(),
		)
		return
	}
}
