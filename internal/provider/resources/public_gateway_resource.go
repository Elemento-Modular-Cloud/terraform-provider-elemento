package resources

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &publicGatewayResource{}
	_ resource.ResourceWithConfigure = &publicGatewayResource{}
)

func NewPublicGatewayResource() resource.Resource {
	return &publicGatewayResource{}
}

type publicGatewayResource struct {
	common.BaseResource
}


func (r *publicGatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_publicgateway"
}

func (r *publicGatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["network_uuid"] = schema.StringAttribute{
		Required:    true,
		Description: "The UUID of the private network associated with the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the VPC associated with the resource.",
	}
	attributes["allocation_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The allocation ID of the resource.",
	}
	attributes["nat_gateway_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the nat gateway of the resource.",
	}
	attributes["public_subnet_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the public subnet associated with the resource.",
	}
	attributes["egress_enabled"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether egress is enabled for this public gateway.",
	}
	attributes["bytes_out_month"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The traffic (in bytes) sustained last month by the public gateway.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *publicGatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.PublicGatewayResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	publicGatewayResource := client.PublicGatewayRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:        plan.Name.ValueString(),
		NetworkUUID: plan.NetworkUUID.ValueString(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(publicGatewayResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "PublicGateway", err) {
		return
	}

	response, err := r.Client.CreateResource(publicGatewayResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating PublicGateway Resource",
			"Could not create publicgateway resource, unexpected error: "+err.Error(),
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
				"Error Reading PublicGateway Resource Status",
				"Could not read publicgateway resource status: "+err.Error(),
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
			fmt.Sprintf("CreatePublicGatewayResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.VPCID = types.StringNull()
	plan.NATGatewayID = types.StringNull()
	plan.AllocationID = types.StringNull()
	plan.PublicSubnetID = types.StringNull()
	plan.EgressEnabled = types.BoolNull()
	plan.BytesOutMonth = types.Int64Null()
	plan.State = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *publicGatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.PublicGatewayResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = model.Read(ctx, r.Client, common.RemoveResource(resp))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.PaymentURL = types.StringNull()

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

func (r *publicGatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *publicGatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.PublicGatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusCode, err := r.Client.DeleteResource(state.UUID.ValueString(), "publicgateway")
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting PublicGateway Resource",
				"Could not delete publicgateway resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting PublicGateway Resource",
				"Could not delete publicgateway resource, unexpected error: "+err.Error(),
			)
		}
		return
	}
}
