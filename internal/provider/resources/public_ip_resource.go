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
	_ resource.Resource              = &publicIPResource{}
	_ resource.ResourceWithConfigure = &publicIPResource{}
)

func NewPublicIPResource() resource.Resource {
	return &publicIPResource{}
}

type publicIPResource struct {
	common.BaseResource
}

func (r *publicIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_publicip"
}

func (r *publicIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["public_ip"] = schema.StringAttribute{
		Computed:    true,
		Description: "The public IP of the resource.",
	}
	attributes["allocation_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The allocation ID of the resource.",
	}
	attributes["domain"] = schema.StringAttribute{
		Computed:    true,
		Description: "The domain of the resource.",
	}
	attributes["association_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The association ID of the resource.",
	}
	attributes["associated_instance_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the instance associated with the resource.",
	}
	attributes["associated_network_interface_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the network interface associated with the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *publicIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.PublicIPResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	publicIPResource := client.PublicIPRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name: plan.Name.ValueString(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(publicIPResource)
	common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "PublicIP", err)

	response, err := r.Client.CreateResource(publicIPResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating PublicIP Resource",
			"Could not create publicip resource, unexpected error: "+err.Error(),
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
				"Error Reading PublicIP Resource Status",
				"Could not read publicip resource status: "+err.Error(),
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
			fmt.Sprintf("CreatePublicIPResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.PublicIP = types.StringNull()
	plan.AllocationID = types.StringNull()
	plan.Domain = types.StringNull()
	plan.AssociationID = types.StringNull()
	plan.AssociatedInstanceID = types.StringNull()
	plan.AssociatedNetworkInterfaceID = types.StringNull()
	plan.State = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *publicIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.PublicIPResourceModel
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

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *publicIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *publicIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.PublicIPResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.DeleteResource(state.UUID.ValueString(), "publicip")
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting PublicIP Resource",
			"Could not delete publicip resource, unexpected error: "+err.Error(),
		)
		return
	}
}
