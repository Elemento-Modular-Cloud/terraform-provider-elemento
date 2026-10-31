package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
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
	baseResource
}

type loadBalancerResourceModel struct {
	baseResourceModel
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	Port            types.Int64  `tfsdk:"port"`
	Scheme          types.String `tfsdk:"scheme"`
	DNSName         types.String `tfsdk:"dns_name"`
	ARN             types.String `tfsdk:"arn"`
	TargetGroupARN  types.String `tfsdk:"target_group_arn"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	VPCID           types.String `tfsdk:"vpc_id"`
	State           types.String `tfsdk:"state"`
}

func (r *loadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

func (r *loadBalancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := baseSchema()

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
	var plan loadBalancerResourceModel
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

	canCreateResponse, err := r.client.CanCreateResource(loadBalancerResource)
	ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "LoadBalancer", err)

	response, err := r.client.CreateResource(loadBalancerResource)
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

		status, err := r.client.GetResourceStatus(res.BillingUUID)
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
	var state loadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resource *client.LoadBalancerResponse
	if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.LoadBalancerResponse](r.client, "loadbalancer")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading LoadBalancer Resources",
				"Could not list loadbalancer resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.LoadBalancerResponse](r.client, state.UUID.ValueString(), "loadbalancer")
		if err != nil && code != 404 {
			resp.Diagnostics.AddError(
				"Error Reading LoadBalancer Resource"+string(code),
				"Could not read loadbalancer resource: "+err.Error(),
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
	state.Type = types.StringValue(resource.Type)
	state.IsPublic = types.BoolValue(resource.IsPublic)
	state.Port = types.Int64Value(resource.Port)
	state.Scheme = types.StringValue(resource.Scheme)
	state.DNSName = types.StringValue(resource.DNSName)
	state.ARN = types.StringValue(resource.ARN)
	state.TargetGroupARN = types.StringValue(resource.TargetGroupARN)
	state.SecurityGroupID = types.StringValue(resource.SecurityGroupID)
	state.VPCID = types.StringValue(resource.VPCID)

	state.State = types.StringValue(resource.State)
	state.PaymentURL = types.StringNull()

	// // NOTE: LoadBalancer credentials contain no useful extra information
	// credentials, err := r.client.GetLoadBalancerCredentials(resource.ServiceUUID)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Reading LoadBalancer Credentials",
	// 		"Could not read loadbalancer credentials: "+err.Error(),
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

func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResource(state.UUID.ValueString(), "loadbalancer")
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting LoadBalancer Resource",
			"Could not delete loadbalancer resource, unexpected error: "+err.Error(),
		)
		return
	}
}
