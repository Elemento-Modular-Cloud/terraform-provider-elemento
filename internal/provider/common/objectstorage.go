package common

import (
	"context"
	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ObjectStorageDataSourceModel struct {
	BaseDataSourceModel
	Name            types.String `tfsdk:"name"`
	ActiveStorage   types.Int32  `tfsdk:"active_storage"`
	Endpoint        types.String `tfsdk:"endpoint"`
	AccountName     types.String `tfsdk:"account_name"`
	ContainerName   types.String `tfsdk:"container_name"`
	SasToken        types.String `tfsdk:"sas_token"`
	AzureCliCommand types.String `tfsdk:"azure_cli_command"`
}

type ObjectStorageResourceModel struct {
	ObjectStorageDataSourceModel
	ResourceExtraModel
	SizeTB types.Int64 `tfsdk:"size_tb"`
}

func (m *ObjectStorageDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
    var resource *client.ObjectStorageResponse
    if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
        resources, err := c.ListObjectStorageResources(ctx)
        if err != nil {
            diags.AddError(
                "Error Reading ObjectStorage Resources",
                "Could not list objectstorage resources: "+err.Error(),
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
        res, code, err := c.GetObjectStorageResource(ctx, m.UUID.ValueString())
        if err != nil && code != 404 {
            diags.AddError(
                "Error Reading ObjectStorage Resource",
                "Could not read objectstorage resource: "+err.Error(),
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
				"Error Reading ObjectStorage Resource Status",
				"Could not read objectstorage resource status: "+err.Error(),
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

    m.Status        = types.StringPointerValue(status)
    m.Name          = types.StringValue(resource.Name)
    m.ActiveStorage = types.Int32Value(int32(resource.ActiveStorage))
    m.Endpoint      = types.StringValue(resource.Endpoint)
    m.Region        = types.StringValue(resource.Region)
    m.UUID          = types.StringValue(resource.UUID)

    credentials, err := c.GetObjectStorageCredentials(ctx, resource.UUID)
    if err != nil {
        diags.AddWarning(
            "Error Reading ObjectStorage Credentials",
            "Could not read objectstorage credentials: "+err.Error(),
        )
        return true, diags
    }

    m.AccountName = types.StringValue(credentials.AccountName)
    m.ContainerName = types.StringValue(credentials.ContainerName)
    m.SasToken = types.StringValue(credentials.SasToken)
    m.AzureCliCommand = types.StringValue(credentials.AzureCliCommand)

    return true, diags
}
