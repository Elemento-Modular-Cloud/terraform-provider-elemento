package common

import (
	"context"
	"strconv"
	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RegistryDataSourceModel struct {
	BaseDataSourceModel
	Name            types.String `tfsdk:"name"`
	SizeGB          types.Int64  `tfsdk:"size_gb"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	State           types.String `tfsdk:"state"`
	RepositoryURI   types.String `tfsdk:"repository_uri"`
	RegistryID      types.String `tfsdk:"registry_id"`
	ImageCount      types.Int64  `tfsdk:"image_count"`
	UsedBytes       types.Int64  `tfsdk:"used_bytes"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	RegistryType    types.String `tfsdk:"registry_type"`
	LoginCommand    types.String `tfsdk:"login_command"`
}

type RegistryResourceModel struct {
	RegistryDataSourceModel
	ResourceExtraModel
}

func (m *RegistryDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var resource *client.RegistryResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.RegistryResponse](c, ctx, "registry")
		if err != nil {
			diags.AddError(
				"Error Reading Registry Resources",
				"Could not list registry resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.RegistryResponse](c, ctx, m.UUID.ValueString(), "registry")
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading Registry Resource"+strconv.Itoa(code),
				"Could not read registry resource: "+err.Error(),
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
				"Error Reading Registry Resource Status",
				"Could not read registry resource status: "+err.Error(),
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
	m.UUID = types.StringValue(resource.ServiceUUID)
	m.CreatorUUID = types.StringValue(resource.CreatorUUID)
	m.CreationDate = types.StringValue(resource.CreationDate)
	m.SizeGB = types.Int64Value(resource.SizeGB)
	m.IsPublic = types.BoolValue(resource.IsPublic)
	m.RepositoryURI = types.StringValue(resource.RepositoryURI)
	m.RegistryID = types.StringValue(resource.RegistryID)
	m.ImageCount = types.Int64Value(resource.ImageCount)
	m.UsedBytes = types.Int64Value(resource.UsedBytes)
	m.State = types.StringValue(resource.State)

	credentials, err := c.GetRegistryCredentials(ctx, resource.ServiceUUID)
	if err != nil {
		diags.AddWarning(
			"Error Reading Registry Credentials",
			"Could not read registry credentials: "+err.Error(),
		)
		return true, diags
	}

	m.AccessKeyID = types.StringValue(credentials.AccessKeyID)
	m.SecretAccessKey = types.StringValue(credentials.SecretAccessKey)
	m.RegistryType = types.StringValue(credentials.RegistryType)
	m.LoginCommand = types.StringValue(credentials.LoginCommand)

	return true, diags
}
