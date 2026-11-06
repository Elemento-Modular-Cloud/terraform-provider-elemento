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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &serverlessResource{}
	_ resource.ResourceWithConfigure = &serverlessResource{}
)

func NewServerlessResource() resource.Resource {
	return &serverlessResource{}
}

type serverlessResource struct {
	common.BaseResource
}

type serverlessResourceModel struct {
	common.BaseResourceModel
	Name            types.String `tfsdk:"name"`
	Runtime         types.String `tfsdk:"runtime"`
	Handler         types.String `tfsdk:"handler"`
	Memory          types.Int64  `tfsdk:"memory"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	Environment     types.Map    `tfsdk:"environment"`
	CodeZipB64      types.String `tfsdk:"code_zip_b64"`
	URL             types.String `tfsdk:"url"`
	ARN             types.String `tfsdk:"arn"`
	State           types.String `tfsdk:"state"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	AuthType        types.String `tfsdk:"auth_type"`
}

func (r *serverlessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serverless"
}

func (r *serverlessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["runtime"] = schema.StringAttribute{
		Required:    true,
		Description: "The runtime to use for the serverless resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["handler"] = schema.StringAttribute{
		Required:    true,
		Description: "The handler to use for the serverless resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["memory"] = schema.Int64Attribute{
		Required:    true,
		Description: "The memory to allocate for the serverless resource.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["timeout"] = schema.Int64Attribute{
		Required:    true,
		Description: "The timeout for the serverless resource.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["is_public"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether the resource is public or not.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["environment"] = schema.MapAttribute{
		Required:    true,
		ElementType: types.StringType,
		Description: "The environment variables for the serverless resource.",
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.RequiresReplace(),
		},
	}
	attributes["code_zip_b64"] = schema.StringAttribute{
		Required:    true,
		Description: "The base64-encoded zip to use for the serverless resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["url"] = schema.StringAttribute{
		Computed:    true,
		Description: "The URL of the resource.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ARN of the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The access key ID for the resource.",
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "The secret access key of the resource.",
	}
	attributes["auth_type"] = schema.StringAttribute{
		Computed:    true,
		Description: "The type of auth used by the resource.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *serverlessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverlessResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var environment map[string]string
	diags = plan.Environment.ElementsAs(ctx, &environment, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverlessResource := client.ServerlessRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:        plan.Name.ValueString(),
		Runtime:     plan.Runtime.ValueString(),
		Handler:     plan.Handler.ValueString(),
		Memory:      plan.Memory.ValueInt64(),
		Timeout:     plan.Timeout.ValueInt64(),
		IsPublic:    plan.IsPublic.ValueBool(),
		Environment: environment,
		CodeZipB64:  plan.CodeZipB64.ValueString(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(serverlessResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "Serverless", err) {
		return
	}

	response, err := r.Client.CreateResource(serverlessResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Serverless Resource",
			"Could not create serverless resource, unexpected error: "+err.Error(),
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
				"Error Reading Serverless Resource Status",
				"Could not read serverless resource status: "+err.Error(),
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
			fmt.Sprintf("CreateServerlessResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.URL = types.StringNull()
	plan.ARN = types.StringNull()
	plan.State = types.StringNull()
	plan.AccessKeyID = types.StringNull()
	plan.SecretAccessKey = types.StringNull()
	plan.AuthType = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *serverlessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverlessResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resource *client.ServerlessResponse
	if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.ServerlessResponse](r.Client, "serverless")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Serverless Resources",
				"Could not list serverless resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.ServerlessResponse](r.Client, state.UUID.ValueString(), "serverless")
		if err != nil && code != 404 {
			resp.Diagnostics.AddError(
				"Error Reading Serverless Resource"+string(code),
				"Could not read serverless resource: "+err.Error(),
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
	state.Runtime = types.StringValue(resource.Runtime)
	state.Handler = types.StringValue(resource.Handler)
	state.Memory = types.Int64Value(resource.Memory)
	state.Timeout = types.Int64Value(resource.Timeout)
	state.IsPublic = types.BoolValue(resource.IsPublic)
	state.URL = types.StringValue(resource.URL)
	state.ARN = types.StringValue(resource.ARN)
	state.State = types.StringValue(resource.State)
	state.PaymentURL = types.StringNull()

	credentials, err := r.Client.GetServerlessCredentials(resource.ServiceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Serverless Credentials",
			"Could not read serverless credentials: "+err.Error(),
		)
		return
	}

	state.AccessKeyID = types.StringValue(credentials.AccessKeyID)
	state.SecretAccessKey = types.StringValue(credentials.SecretAccessKey)
	state.AuthType = types.StringValue(credentials.AuthType)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *serverlessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *serverlessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverlessResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.DeleteResource(state.UUID.ValueString(), "serverless")
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting Serverless Resource",
			"Could not delete serverless resource, unexpected error: "+err.Error(),
		)
		return
	}
}
