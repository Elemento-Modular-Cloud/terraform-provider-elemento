package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OpenClawDataSourceModel struct {
	BaseDataSourceModel
	VMName     types.String `tfsdk:"vm_name"`
	CPUCores   types.Int32  `tfsdk:"cpu_cores"`
	RAMSizeMB  types.Int32  `tfsdk:"ram_size_mb"`
	DiskSizeGB types.Int64  `tfsdk:"disk_size_gb"`
	State      types.String `tfsdk:"state"`
	ServiceURL types.String `tfsdk:"service_url"`
	Token      types.String `tfsdk:"token"`
}

type OpenClawResourceModel struct {
	OpenClawDataSourceModel
	ResourceExtraModel
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	SSHKey   types.String `tfsdk:"ssh_key"`
}

func (m *OpenClawDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound func(ctx context.Context)) diag.Diagnostics {
	var diags diag.Diagnostics

	resources, err := c.ListComputeResources()
	if err != nil {
		diags.AddError(
			"Error Reading OpenClaw Resources",
			"Could not list openclaw resources: "+err.Error(),
		)
		return diags
	}

	var resource *client.ComputeListItem
	for _, res := range resources {
		if res.BillingUUID == m.BillingUUID.ValueString() || res.UUID == m.UUID.ValueString() {
			resource = &res
			break
		}
	}

	billing_uuid := m.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		status, err = c.GetResourceStatus(m.BillingUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading OpenClaw Resource Status",
				"Could not read openclaw resource status: "+err.Error(),
			)
			return diags
		}
	} else {
		status = nil
	}

	if resource == nil && (status == nil || *status == "deleted" || *status == "error" || *status == "canceled") {
		onNotFound(ctx)
		return diags
	}

	m.Status = types.StringPointerValue(status)
	m.VMName = types.StringValue(resource.Requirements.VMName)
	m.CPUCores = types.Int32Value(int32(resource.Requirements.Slots))
	m.RAMSizeMB = types.Int32Value(int32(resource.Requirements.RamSize))
	m.VMName = types.StringValue(resource.Requirements.VMName)
	m.Region = types.StringValue(resource.Region)
	m.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	m.UUID = types.StringValue(resource.UUID)

	credentials, err := c.GetOpenClawCredentials(resource.UUID)
	if err != nil {
		diags.AddError(
			"Error Reading OpenClaw Credentials",
			"Could not read openclaw credentials: "+err.Error(),
		)
		return diags
	}

	m.State = types.StringValue(credentials.State)
	m.ServiceURL = types.StringValue(credentials.URL)
	m.Token = types.StringValue(credentials.Token)

	return diags
}
