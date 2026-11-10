package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FileStorageDataSourceModel struct {
	BaseDataSourceModel
	Name         types.String  `tfsdk:"name"`
	SizeGB       types.Int64   `tfsdk:"size_gb"`
	UsedGB       types.Float64 `tfsdk:"used_gb"`
	DNSName      types.String  `tfsdk:"dns_name"`
	State        types.String  `tfsdk:"state"`
	NFSHost      types.String  `tfsdk:"nfs_host"`
	NFSPath      types.String  `tfsdk:"nfs_path"`
	NFSVersion   types.String  `tfsdk:"nfs_version"`
	MountCommand types.String  `tfsdk:"mount_command"`
}

type FileStorageResourceModel struct {
	FileStorageDataSourceModel
	ResourceExtraModel
	IsPublic types.Bool `tfsdk:"is_public"`
}

func (m *FileStorageDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) diag.Diagnostics {
	var diags diag.Diagnostics

	var resource *client.FileStorageResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := c.ListFileStorageResources()
		if err != nil {
			diags.AddError(
				"Error Reading FileStorage Resources",
				"Could not list filestorage resources: "+err.Error(),
			)
			return diags
		}

		for _, res := range resources {
			if res.BillingUUID == m.BillingUUID.ValueString() {
				resource = &res
				break
			}
		}
	} else {
		res, code, err := c.GetFileStorageResource(m.UUID.ValueString())
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading FileStorage Resource"+string(code),
				"Could not read filestorage resource: "+err.Error(),
			)
			return diags
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

	m.Status = types.StringPointerValue(status)
	m.Name = types.StringValue(resource.Name)
	m.Region = types.StringValue(resource.Region)
	m.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	m.UUID = types.StringValue(resource.ServiceUUID)
	m.CreatorUUID = types.StringValue(resource.CreatorUUID)
	m.CreationDate = types.StringValue(resource.CreationDate)
	m.SizeGB = types.Int64Value(resource.SizeGB)
	m.UsedGB = types.Float64Value(resource.UsedGB)
	m.DNSName = types.StringValue(resource.DNSName)
	m.State = types.StringValue(resource.State)

	credentials, err := c.GetFileStorageCredentials(resource.ServiceUUID)
	if err != nil {
		diags.AddError(
			"Error Reading FileStorage Credentials",
			"Could not read filestorage credentials: "+err.Error(),
		)
		return diags
	}

	m.NFSHost = types.StringValue(credentials.NFSHost)
	m.NFSPath = types.StringValue(credentials.NFSPath)
	m.NFSVersion = types.StringValue(credentials.NFSVersion)
	m.MountCommand = types.StringValue(credentials.MountCommand)

	return diags
}
