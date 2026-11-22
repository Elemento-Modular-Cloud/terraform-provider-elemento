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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &publicIPResource{}
	_ resource.ResourceWithConfigure   = &publicIPResource{}
	_ resource.ResourceWithImportState = &publicIPResource{}
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

func (r *publicIPResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema(ctx)

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "Name of the public IP.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["public_ip"] = schema.StringAttribute{
		Computed:    true,
		Description: "The IP address.",
	}
	attributes["allocation_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "Allocation ID of the address.",
	}
	attributes["domain"] = schema.StringAttribute{
		Computed:    true,
		Description: "Domain of the address.",
	}
	attributes["association_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the association with an instance.",
	}
	attributes["associated_instance_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the instance the address is attached to.",
	}
	attributes["associated_network_interface_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the network interface the address is attached to.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
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

	createTimeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
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

	canCreateResponse, err := r.Client.CanCreateResource(ctx, publicIPResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "PublicIP", err) {
		return
	}

	response, err := r.Client.CreateResource(ctx, publicIPResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating PublicIP Resource",
			"Could not create publicip resource, unexpected error: "+err.Error(),
		)
		return
	}

	mustRead := false
	switch res := response.(type) {
	case client.PublicCreateResponse:
		plan.BillingUUID = types.StringValue(res.BillingUUID)
		plan.PaymentURL = types.StringValue(res.PaymentURL)
		plan.UUID = types.StringNull()

		status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Reading PublicIP Resource Status",
				"Could not read publicip resource status: "+err.Error(),
			)
			status = nil
		}

		plan.Status = types.StringPointerValue(status)
	case client.PrivateCreateResponse:
		plan.BillingUUID = types.StringNull()
		plan.PaymentURL = types.StringNull()
		plan.UUID = types.StringValue(res.ServiceUUID)
		plan.Status = types.StringValue(res.Status)

		mustRead = true
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

	if mustRead {
    	stateConf := common.GetCreateStateConf(ctx, r.Client, createTimeout, &plan)

		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			resp.Diagnostics.AddError("Timeout waiting for resource creation", err.Error())
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *publicIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.PublicIPResourceModel
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

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error Deleting PublicIP Resource",
			"Could not delete publicip resource: resource uuid is null.",
		)
		return
	}

	statusCode, err := r.Client.DeleteResource(ctx, state.UUID.ValueString(), "publicip")
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting PublicIP Resource",
				"Could not delete publicip resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting PublicIP Resource",
				"Could not delete publicip resource, unexpected error: "+err.Error(),
			)
		}
		return
	}
}

func (r *publicIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
