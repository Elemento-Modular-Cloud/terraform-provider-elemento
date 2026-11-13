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
)

var (
	_ resource.Resource              = &privateNetworkResource{}
	_ resource.ResourceWithConfigure = &privateNetworkResource{}
	_ resource.ResourceWithImportState = &privateNetworkResource{}
)

func NewPrivateNetworkResource() resource.Resource {
	return &privateNetworkResource{}
}

type privateNetworkResource struct {
	common.BaseResource
}

func (r *privateNetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privatenetwork"
}

func (r *privateNetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

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
		Description: "The state of the resource.",
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

func (r *privateNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.PrivateNetworkResourceModel
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

	canCreateResponse, err := r.Client.CanCreatePrivateNetworkResource(ctx, privateNetworkResource)
    if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "PrivateNetwork", err) {
	    return
    }

	response, err := r.Client.CreatePrivateNetworkResource(ctx, privateNetworkResource)
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

		diags = resp.State.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)

		status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
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
	plan.Subnets = types.ListNull(common.PrivateNetworkSubnetObjectType())
	plan.State = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *privateNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.PrivateNetworkResourceModel
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

func (r *privateNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *privateNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.PrivateNetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error deleting Resource",
			"Could not delete resource: resource uuid is null.",
		)
		return
	}

	statusCode, err := r.Client.DeletePrivateNetworkResource(ctx, state.UUID.ValueString())
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting PrivateNetwork Resource",
				"Could not delete privatenetwork resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting PrivateNetwork Resource",
				"Could not delete privatenetwork resource, unexpected error: "+err.Error(),
			)
		}
		return
	}
}

func (r *privateNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
