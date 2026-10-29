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
	_ resource.Resource              = &publicIPResource{}
	_ resource.ResourceWithConfigure = &publicIPResource{}
)

func NewPublicIPResource() resource.Resource {
	return &publicIPResource{}
}

type publicIPResource struct {
	baseResource
}

type publicIPResourceModel struct {
	baseResourceModel
	Name                         types.String `tfsdk:"name"`
	PublicIP                     types.String `tfsdk:"public_ip"`
	AllocationID                 types.String `tfsdk:"allocation_id"`
	Domain                       types.String `tfsdk:"domain"`
	AssociationID                types.String `tfsdk:"association_id"`
	AssociatedInstanceID         types.String `tfsdk:"associated_instance_id"`
	AssociatedNetworkInterfaceID types.String `tfsdk:"associated_network_interface_id"`
	State                        types.String `tfsdk:"state"`
}

func (r *publicIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_publicip"
}

func (r *publicIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := baseSchema()

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
		Description: "The name of the DNS associated with the resource.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *publicIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan publicIPResourceModel
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

	canCreateResponse, err := r.client.CanCreateResource(publicIPResource)
	ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "PublicIP", err)

	response, err := r.client.CreateResource(publicIPResource)
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

		status, err := r.client.GetResourceStatus(res.BillingUUID)
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
	var state publicIPResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resource *client.PublicIPResponse
	if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.PublicIPResponse](r.client, "publicip")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading PublicIP Resources",
				"Could not list publicip resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.PublicIPResponse](r.client, state.UUID.ValueString(), "publicip")
		if err != nil && code != 404 {
			resp.Diagnostics.AddError(
				"Error Reading PublicIP Resource"+string(code),
				"Could not read publicip resource: "+err.Error(),
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
	state.PublicIP = types.StringValue(resource.PublicIP)
	state.AllocationID = types.StringValue(resource.AllocationID)
	state.Domain = types.StringValue(resource.Domain)
	state.AssociationID = types.StringValue(resource.AssociationID)
	state.AssociatedInstanceID = types.StringValue(resource.AssociatedInstanceID)
	state.AssociatedNetworkInterfaceID = types.StringValue(resource.AssociatedNetworkInterfaceID)
	state.State = types.StringValue(resource.State)
	state.PaymentURL = types.StringNull()

	// // NOTE: PublicIP credentials contain no useful extra information
	// credentials, err := r.client.GetPublicIPCredentials(resource.ServiceUUID)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Reading PublicIP Credentials",
	// 		"Could not read publicip credentials: "+err.Error(),
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

func (r *publicIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *publicIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state publicIPResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResource(state.UUID.ValueString(), "publicip")
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting PublicIP Resource",
			"Could not delete publicip resource, unexpected error: "+err.Error(),
		)
		return
	}
}
