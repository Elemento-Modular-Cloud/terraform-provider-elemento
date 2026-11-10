package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PublicIPDataSourceModel struct {
	BaseDataSourceModel
	Name                         types.String `tfsdk:"name"`
	PublicIP                     types.String `tfsdk:"public_ip"`
	AllocationID                 types.String `tfsdk:"allocation_id"`
	Domain                       types.String `tfsdk:"domain"`
	AssociationID                types.String `tfsdk:"association_id"`
	AssociatedInstanceID         types.String `tfsdk:"associated_instance_id"`
	AssociatedNetworkInterfaceID types.String `tfsdk:"associated_network_interface_id"`
	State                        types.String `tfsdk:"state"`
}

type PublicIPResourceModel struct {
	PublicIPDataSourceModel
	ResourceExtraModel
}

func (model *PublicIPDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) diag.Diagnostics {
	var diags diag.Diagnostics

	var resource *client.PublicIPResponse
	if model.UUID.IsNull() || model.UUID.IsUnknown() || model.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.PublicIPResponse](c, "publicip")
		if err != nil {
			diags.AddError(
				"Error Reading PublicIP Resources",
				"Could not list publicip resources: "+err.Error(),
			)
			return diags
		}

		for _, res := range resources {
			if res.BillingUUID == model.BillingUUID.ValueString() {
				resource = &res
				break
			}
		}
	} else {
		res, code, err := client.GetResource[client.PublicIPResponse](c, model.UUID.ValueString(), "publicip")
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading PublicIP Resource"+string(code),
				"Could not read publicip resource: "+err.Error(),
			)
			return diags
		}
		resource = res
	}

	billing_uuid := model.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		var err error
		status, err = c.GetResourceStatus(model.BillingUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading Compute Resource Status",
				"Could not read compute resource status: "+err.Error(),
			)
			return diags
		}
	} else {
		status = nil
	}

	if resource == nil {
		statusDeleted := status == nil || *status == "deleted" || *status == "error" || *status == "canceled"
		onNotFound(ctx, statusDeleted)
		return diags
	}

	model.Status = types.StringPointerValue(status)
	model.Name = types.StringValue(resource.Name)
	model.Region = types.StringValue(resource.Region)
	model.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	model.UUID = types.StringValue(resource.ServiceUUID)
	model.CreatorUUID = types.StringValue(resource.CreatorUUID)
	model.CreationDate = types.StringValue(resource.CreationDate)
	model.PublicIP = types.StringValue(resource.PublicIP)
	model.AllocationID = types.StringValue(resource.AllocationID)
	model.Domain = types.StringValue(resource.Domain)
	model.AssociationID = types.StringValue(resource.AssociationID)
	model.AssociatedInstanceID = types.StringValue(resource.AssociatedInstanceID)
	model.AssociatedNetworkInterfaceID = types.StringValue(resource.AssociatedNetworkInterfaceID)
	model.State = types.StringValue(resource.State)

	// // NOTE: PublicIP credentials contain no useful extra information
	// credentials, err := c.GetPublicIPCredentials(resource.ServiceUUID)
	// if err != nil {
	// 	diags.AddError(
	// 		"Error Reading PublicIP Credentials",
	// 		"Could not read publicip credentials: "+err.Error(),
	// 	)
	// 	return diags
	// }

	// state.NFSHost = types.StringValue(credentials.NFSHost)
	// state.NFSPath = types.StringValue(credentials.NFSPath)
	// state.NFSVersion = types.StringValue(credentials.NFSVersion)
	// state.MountCommand = types.StringValue(credentials.MountCommand)

	return diags
}
