package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
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
	baseResource
}

type publicGatewayResourceModel struct {
	baseResourceModel
	Name           types.String `tfsdk:"name"`
	NetworkUUID    types.String `tfsdk:"network_uuid"`
	VPCID          types.String `tfsdk:"vpc_id"`
	NATGatewayID   types.String `tfsdk:"nat_gateway_id"`
	AllocationID   types.String `tfsdk:"allocation_id"`
	PublicSubnetID types.String `tfsdk:"public_subnet_id"`
	EgressEnabled  types.Bool   `tfsdk:"egress_enabled"`
	BytesOutMonth  types.Int64  `tfsdk:"bytes_out_month"`
	State          types.String `tfsdk:"state"`
}

func (r *publicGatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_publicgateway"
}

func (r *publicGatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := baseSchema()

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
	var plan publicGatewayResourceModel
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

	canCreateResponse, err := r.client.CanCreateResource(publicGatewayResource)
	ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "PublicGateway", err)

	response, err := r.client.CreateResource(publicGatewayResource)
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

		status, err := r.client.GetResourceStatus(res.BillingUUID)
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
	var state publicGatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resource *client.PublicGatewayResponse
	if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.PublicGatewayResponse](r.client, "publicgateway")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading PublicGateway Resources",
				"Could not list publicgateway resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.PublicGatewayResponse](r.client, state.UUID.ValueString(), "publicgateway")
		if err != nil && code != 404 {
			resp.Diagnostics.AddError(
				"Error Reading PublicGateway Resource"+string(code),
				"Could not read publicgateway resource: "+err.Error(),
			)
			return
		}
		resource = res
	}

	billing_uuid := state.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		var err error
		status, err = r.client.GetResourceStatus(state.BillingUUID.ValueString())
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
	state.VPCID = types.StringValue(resource.VPCID)
	state.AllocationID = types.StringValue(resource.AllocationID)
	state.PublicSubnetID = types.StringValue(resource.PublicSubnetID)
	state.EgressEnabled = types.BoolValue(resource.EgressEnabled)
	state.BytesOutMonth = types.Int64Value(resource.BytesOutMonth)
	state.State = types.StringValue(resource.State)
	state.PaymentURL = types.StringNull()

	// // NOTE: PublicGateway credentials contain no useful extra information
	// credentials, err := r.client.GetPublicGatewayCredentials(resource.ServiceUUID)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Reading PublicGateway Credentials",
	// 		"Could not read publicgateway credentials: "+err.Error(),
	// 	)
	// 	return
	// }

	// state.NFSHost = types.StringValue(credentials.NFSHost)
	// state.NFSPath = types.StringValue(credentials.NFSPath)
	// state.NFSVersion = types.StringValue(credentials.NFSVersion)
	// state.MountCommand = types.StringValue(credentials.MountCommand)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *publicGatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *publicGatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state publicGatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResource(state.UUID.ValueString(), "publicgateway")
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting PublicGateway Resource",
			"Could not delete publicgateway resource, unexpected error: "+err.Error(),
		)
		return
	}
}
