package common

import (
	"context"
	"strconv"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivateNetworkDataSourceModel struct {
	BaseDataSourceModel
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

func PrivateNetworkSubnetObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"az":        types.StringType,
			"subnet_id": types.StringType,
			"cidr":      types.StringType,
		},
	}
}

type PrivateNetworkResourceModel struct {
	PrivateNetworkDataSourceModel
	ResourceExtraModel
}

func (m *PrivateNetworkDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var resource *client.PrivateNetworkResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := c.ListPrivateNetworkResources()
		if err != nil {
			diags.AddError(
				"Error Reading PrivateNetwork Resources",
				"Could not list privatenetwork resources: "+err.Error(),
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
		res, code, err := c.GetPrivateNetworkResource(m.UUID.ValueString())
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading PrivateNetwork Resource"+strconv.Itoa(code),
				"Could not read privatenetwork resource: "+err.Error(),
			)
			return true, diags
		}
		resource = res
	}

	billing_uuid := m.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		var err error
		status, err = c.GetResourceStatus(m.BillingUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading PrivateNetwork Resource Status",
				"Could not read privatenetwork resource status: "+err.Error(),
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
	m.CIDR = types.StringValue(resource.CIDR)
	m.AutoAssignCIDR = types.BoolValue(resource.AutoAssignCIDR)
	m.VPCID = types.StringValue(resource.VPCID)
	m.State = types.StringValue(resource.State)

	var subnets []subnetModel
	for _, subnet := range resource.Subnets {
		subnets = append(subnets, subnetModel{
			AZ:   types.StringValue(subnet.AZ),
			CIDR: types.StringValue(subnet.CIDR),
			ID:   types.StringValue(subnet.ID),
		})
	}

	var tmpDiags diag.Diagnostics
	m.Subnets, tmpDiags = types.ListValueFrom(ctx, PrivateNetworkSubnetObjectType(), subnets)
	diags.Append(tmpDiags...)

	// // NOTE: current network credentials are not useful so we ignore them.
	// credentials, err := c.GetPrivateNetworkCredentials(resource.ServiceUUID)
	// if err != nil {
	// 	diags.AddWarning(
	// 		"Error Reading PrivateNetwork Credentials",
	// 		"Could not read privatenetwork credentials: "+err.Error(),
	// 	)
	// 	return true, diags
	// }


	return true, diags
}
