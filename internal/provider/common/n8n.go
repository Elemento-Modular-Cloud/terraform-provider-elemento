package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type N8NDataSourceModel struct {
	BaseDataSourceModel
	VMName     types.String `tfsdk:"vm_name"`
	CPUCores   types.Int32  `tfsdk:"cpu_cores"`
	RAMSizeMB  types.Int32  `tfsdk:"ram_size_mb"`
	DiskSizeGB types.Int64  `tfsdk:"disk_size_gb"`
	State      types.String `tfsdk:"state"`
	ServiceURL types.String `tfsdk:"service_url"`
}

type N8NResourceModel struct {
	N8NDataSourceModel
	ResourceExtraModel
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	SSHKey   types.String `tfsdk:"ssh_key"`
}

func (m *N8NDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	resources, err := c.ListSpotVMResources(ctx)
	if err != nil {
		diags.AddError(
			"Error Reading n8n Resources",
			"Could not list n8n resources: "+err.Error(),
		)
		return true, diags
	}

	var resource *client.SpotVMListItem
	for _, res := range resources {
		if res.BillingUUID == m.BillingUUID.ValueString() || res.UUID == m.UUID.ValueString() {
			resource = &res
			break
		}
	}

	billing_uuid := m.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		status, err = c.GetResourceStatus(ctx, m.BillingUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading n8n Resource Status",
				"Could not read n8n resource status: "+err.Error(),
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
	m.VMName = types.StringValue(resource.VMName)
	m.CPUCores = types.Int32Value(int32(resource.CPU.Slots))
	m.RAMSizeMB = types.Int32Value(int32(resource.Mem.CapacityMB))
	m.DiskSizeGB = types.Int64Value(int64(resource.Storage.Boot.SizeGB))
	m.Region = types.StringValue(resource.Region)
	m.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	m.UUID = types.StringValue(resource.UUID)

	credentials, err := c.GetN8NCredentials(ctx, resource.UUID)
	if err != nil {
		diags.AddWarning(
			"Error Reading n8n Credentials",
			"Could not read n8n credentials: "+err.Error(),
		)
		return true, diags
	}

	m.State = types.StringValue(credentials.State)
	m.ServiceURL = types.StringValue(credentials.URL)

	return true, diags
}
