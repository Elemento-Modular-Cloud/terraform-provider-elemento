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
	_ resource.Resource              = &loadBalancerResource{}
	_ resource.ResourceWithConfigure = &loadBalancerResource{}
)

func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

type loadBalancerResource struct {
	common.BaseResource
}

func (r *loadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

func (r *loadBalancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["type"] = schema.StringAttribute{
		Required:    true,
		Description: "The type of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["is_public"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether the resource is public or not.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["port"] = schema.Int64Attribute{
		Required:    true,
		Description: "The port of the resource.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["scheme"] = schema.StringAttribute{
		Computed:    true,
		Description: "The scheme of the resource.",
	}
	attributes["dns_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The DNS name of the resource.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "The arn of the resource.",
	}
	attributes["target_group_arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ARN of the target group associated with the resource.",
	}
	attributes["security_group_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the security group associated with the resource.",
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the VPC associated with the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *loadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.LoadBalancerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loadBalancerResource := client.LoadBalancerRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:     plan.Name.ValueString(),
		Type:     plan.Type.ValueString(),
		IsPublic: plan.IsPublic.ValueBool(),
		Port:     plan.Port.ValueInt64(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(loadBalancerResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "LoadBalancer", err) {
		return
	}

	response, err := r.Client.CreateResource(loadBalancerResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating LoadBalancer Resource",
			"Could not create loadbalancer resource, unexpected error: "+err.Error(),
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
				"Error Reading LoadBalancer Resource Status",
				"Could not read loadbalancer resource status: "+err.Error(),
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
			fmt.Sprintf("CreateLoadBalancerResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.Scheme = types.StringNull()
	plan.DNSName = types.StringNull()
	plan.ARN = types.StringNull()
	plan.TargetGroupARN = types.StringNull()
	plan.SecurityGroupID = types.StringNull()
	plan.VPCID = types.StringNull()
	plan.State = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.LoadBalancerResourceModel
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

func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.LoadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusCode, err := r.Client.DeleteResource(state.UUID.ValueString(), "loadbalancer")
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting LoadBalancer Resource",
				"Could not delete loadbalancer resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting LoadBalancer Resource",
				"Could not delete loadbalancer resource, unexpected error: "+err.Error(),
			)
		}
		return
	}
}
