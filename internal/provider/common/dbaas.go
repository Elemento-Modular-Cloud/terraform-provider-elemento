package common

import (
	"context"
	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DBAASDataSourceModel struct {
	BaseDataSourceModel
	Name             types.String `tfsdk:"name"`
	Engine           types.String `tfsdk:"engine"`
	BackupTime       types.String `tfsdk:"backup_time"`
	DiskSizeGB       types.Int64  `tfsdk:"disk_size_gb"`
	NodesNumber      types.Int32  `tfsdk:"nodes_number"`
	DatabaseHost     types.String `tfsdk:"database_host"`
	DatabasePort     types.String `tfsdk:"database_port"`
	DatabaseUser     types.String `tfsdk:"database_user"`
	DatabasePassword types.String `tfsdk:"database_password"`
	DatabaseName     types.String `tfsdk:"database_name"`
	DatabaseSSLMode  types.String `tfsdk:"database_ssl_mode"`
	PublicURI        types.String `tfsdk:"public_uri"`
}

type DBAASResourceModel struct {
	DBAASDataSourceModel
	ResourceExtraModel
	Password types.String `tfsdk:"password"`
}

func (m *DBAASDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	var resource *client.DBAASResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := c.ListDBAASResources(ctx)
		if err != nil {
			diags.AddError(
				"Error Reading DBAAS Resources",
				"Could not list dbaas resources: "+err.Error(),
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
		res, code, err := c.GetDBAASResource(ctx, m.UUID.ValueString())
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading DBAAS Resource",
				"Could not read dbaas resource: "+err.Error(),
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
				"Error Reading DBAAS Resource Status",
				"Could not read dbaas resource status: "+err.Error(),
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
	m.UUID = types.StringValue(resource.UUID)
	m.CreatorUUID = types.StringValue(resource.CreatorUUID)
	m.CreationDate = types.StringValue(resource.CreationDate)
	m.BackupTime = types.StringValue(resource.BackupTime)
	m.DiskSizeGB = types.Int64Value(resource.DiskSizeGB)
	m.NodesNumber = types.Int32Value(resource.NodesNumber)
	m.Engine = types.StringValue(resource.Engine)

	credentials, err := c.GetDBAASCredentials(ctx, resource.UUID)
	if err != nil {
		diags.AddWarning(
			"Error Reading DBAAS Credentials",
			"Could not read dbaas credentials: "+err.Error(),
		)
		return true, diags
	}

	m.DatabaseHost = types.StringValue(credentials.DatabaseHost)
	m.DatabasePort = types.StringValue(credentials.DatabasePort)
	m.DatabaseUser = types.StringValue(credentials.DatabaseUser)
	m.DatabasePassword = types.StringValue(credentials.DatabasePassword)
	m.DatabaseName = types.StringValue(credentials.DatabaseName)
	m.DatabaseSSLMode = types.StringValue(credentials.DatabaseSSLMode)
	m.PublicURI = types.StringValue(credentials.PublicURI)

	return true, diags
}
