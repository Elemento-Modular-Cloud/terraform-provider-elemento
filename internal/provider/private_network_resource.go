package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &privateNetworkResource{}
	_ resource.ResourceWithConfigure = &privateNetworkResource{}
)

func NewPrivateNetworkResource() resource.Resource {
	return &privateNetworkResource{}
}

type privateNetworkResource struct {
	client *client.ElementoClient
}

type privateNetworkResourceModel struct {
	baseResourceModel
	Name           types.String `tfsdk:"name"`
	CIDR           types.String `tfsdk:"cidr"`
	AutoAssignCIDR types.Bool   `tfsdk:"auto_assign_cidr"`
	VPCID          types.String `tfsdk:"vpc_id"`
	State          types.String `tfsdk:"state"`
	Subnets        types.List   `tfsdk:"subnets"`
}

type subnetModel struct {
	AZ   types.String `tfsdk:"az"`
	CIDR types.String `tfsdk:"cidr"`
	ID   types.String `tfsdk:"subnet_id"`
}

func subnetObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"az":        types.StringType,
			"subnet_id": types.StringType,
			"cidr":      types.StringType,
		},
	}
}

func (r *privateNetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privatenetwork"
}

func (r *privateNetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := baseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["cidr"] = schema.StringAttribute{
		Required:    true,
		Description: "The network CIDR of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["auto_assign_cidr"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether the CIDR should be automatically assigned for this resource.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The VPC ID of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the DNS associated with the resource.",
	}
	attributes["subnets"] = schema.ListNestedAttribute{
		Computed:    true,
		Description: "The subnets associated with the resource",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"az": schema.StringAttribute{
					Computed:    true,
					Description: "The availability zone of the subnet",
				},
				"subnet_id": schema.StringAttribute{
					Computed:    true,
					Description: "The ID of the subnet",
				},
				"cidr": schema.StringAttribute{
					Computed:    true,
					Description: "The CIDR of the subnet",
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *privateNetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ElementoClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *ElementoClient, got %T. Please report this issue to the provider developer.", req.ProviderData),
		)

		return
	}
	r.client = client
}

func (r *privateNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan privateNetworkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	privateNetworkResource := client.PrivateNetworkRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:           plan.Name.ValueString(),
		CIDR:           plan.CIDR.ValueString(),
		AutoAssignCIDR: plan.AutoAssignCIDR.ValueBool(),
	}

	canCreateResponse, err := r.client.CanCreatePrivateNetworkResource(privateNetworkResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Checking PrivateNetwork Resource Creation",
			"Could not check if privatenetwork resource can be created, unexpected error: "+err.Error(),
		)
		return
	}

	if !canCreateResponse.CanCreate {
		for field, reason := range canCreateResponse.Reasons {
			resp.Diagnostics.AddError(
				"PrivateNetwork Resource Creation Not Allowed",
				fmt.Sprintf("Cannot create privatenetwork resource due to field '%s' value: %s", field, reason),
			)
		}
		return
	}

	response, err := r.client.CreatePrivateNetworkResource(privateNetworkResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating PrivateNetwork Resource",
			"Could not create privatenetwork resource, unexpected error: "+err.Error(),
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
				"Error Reading PrivateNetwork Resource Status",
				"Could not read privatenetwork resource status: "+err.Error(),
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
			fmt.Sprintf("CreatePrivateNetworkResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.VPCID = types.StringNull()
	plan.Subnets = types.ListNull(subnetObjectType())
	plan.State = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *privateNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state privateNetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resource *client.PrivateNetworkResponse
	if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
		resources, err := r.client.ListPrivateNetworkResources()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading PrivateNetwork Resources",
				"Could not list privatenetwork resources: "+err.Error(),
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
		res, code, err := r.client.GetPrivateNetworkResource(state.UUID.ValueString())
		if err != nil && code != 404 {
			resp.Diagnostics.AddError(
				"Error Reading PrivateNetwork Resource"+string(code),
				"Could not read privatenetwork resource: "+err.Error(),
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
	state.CIDR = types.StringValue(resource.CIDR)
	state.AutoAssignCIDR = types.BoolValue(resource.AutoAssignCIDR)
	state.VPCID = types.StringValue(resource.VPCID)
	state.State = types.StringValue(resource.State)
	state.PaymentURL = types.StringNull()

	var subnets []subnetModel
	for _, subnet := range resource.Subnets {
		subnets = append(subnets, subnetModel{
			AZ:   types.StringValue(subnet.AZ),
			CIDR: types.StringValue(subnet.CIDR),
			ID:   types.StringValue(subnet.ID),
		})
	}
	state.Subnets, diags = types.ListValueFrom(ctx, subnetObjectType(), subnets)
	resp.Diagnostics.Append(diags...)

	// // NOTE: current network credentials are not useful so we ignore them.
	// credentials, err := r.client.GetPrivateNetworkCredentials(resource.ServiceUUID)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Reading PrivateNetwork Credentials",
	// 		"Could not read privatenetwork credentials: "+err.Error(),
	// 	)
	// 	return
	// }

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privateNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *privateNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state privateNetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePrivateNetworkResource(state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting PrivateNetwork Resource",
			"Could not delete privatenetwork resource, unexpected error: "+err.Error(),
		)
		return
	}
}
