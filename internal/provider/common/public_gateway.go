package common

import (
	"context"
	"strconv"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PublicGatewayDataSourceModel struct {
	BaseDataSourceModel
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


type PublicGatewayResourceModel struct {
	PublicGatewayDataSourceModel
	ResourceExtraModel
}

func (m *PublicGatewayDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var resource *client.PublicGatewayResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.PublicGatewayResponse](c, ctx, "publicgateway")
		if err != nil {
			diags.AddError(
				"Error Reading PublicGateway Resources",
				"Could not list publicgateway resources: "+err.Error(),
			)
			return true, diags
		}

		for _, res := range resources {
			if res.BillingUUID == m.BillingUUID.ValueString() {
				resource = &res
				break
			}
		}
	} else {
		res, code, err := client.GetResource[client.PublicGatewayResponse](c, ctx, m.UUID.ValueString(), "publicgateway")
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading PublicGateway Resource"+strconv.Itoa(code),
				"Could not read publicgateway resource: "+err.Error(),
			)
			return true, diags
		}
		resource = res
	}

	billing_uuid := m.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		var err error
		status, err = c.GetResourceStatus(ctx, m.BillingUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading PublicGateway Resource Status",
				"Could not read publicgateway resource status: "+err.Error(),
			)
			return true, diags
		}
	} else {
		status = nil
	}

	if resource == nil {
		statusDeleted := status == nil || *status == "deleted" || *status == "error" || *status == "canceled"
		onNotFound(ctx, statusDeleted)
		return !statusDeleted, diags
	}

	m.Status = types.StringPointerValue(status)
	m.Name = types.StringValue(resource.Name)
	m.Region = types.StringValue(resource.Region)
	m.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	m.UUID = types.StringValue(resource.ServiceUUID)
	m.CreatorUUID = types.StringValue(resource.CreatorUUID)
	m.CreationDate = types.StringValue(resource.CreationDate)
	m.VPCID = types.StringValue(resource.VPCID)
	m.AllocationID = types.StringValue(resource.AllocationID)
	m.PublicSubnetID = types.StringValue(resource.PublicSubnetID)
	m.EgressEnabled = types.BoolValue(resource.EgressEnabled)
	m.BytesOutMonth = types.Int64Value(resource.BytesOutMonth)
	m.NetworkUUID = types.StringValue(resource.NetworkUUID)
	m.NATGatewayID = types.StringValue(resource.NATGatewayID)
	m.State = types.StringValue(resource.State)

	// // NOTE: PublicGateway credentials contain no useful extra information
	// credentials, err := c.GetPublicGatewayCredentials(ctx, resource.ServiceUUID)
	// if err != nil {
	// 	diags.AddWarning(
	// 		"Error Reading PublicGateway Credentials",
	// 		"Could not read publicgateway credentials: "+err.Error(),
	// 	)
	// 	return true, diags
	// }

	// state.NFSHost = types.StringValue(credentials.NFSHost)
	// state.NFSPath = types.StringValue(credentials.NFSPath)
	// state.NFSVersion = types.StringValue(credentials.NFSVersion)
	// state.MountCommand = types.StringValue(credentials.MountCommand)

	return true, diags
}
