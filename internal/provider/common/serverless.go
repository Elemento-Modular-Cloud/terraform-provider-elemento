package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ServerlessDataSourceModel struct {
	BaseDataSourceModel
	Name            types.String `tfsdk:"name"`
	Runtime         types.String `tfsdk:"runtime"`
	Handler         types.String `tfsdk:"handler"`
	Memory          types.Int64  `tfsdk:"memory"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	URL             types.String `tfsdk:"url"`
	ARN             types.String `tfsdk:"arn"`
	State           types.String `tfsdk:"state"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	AuthType        types.String `tfsdk:"auth_type"`
}

type ServerlessResourceModel struct {
	ServerlessDataSourceModel
	ResourceExtraModel
	Environment types.Map    `tfsdk:"environment"`
	CodeZipB64  types.String `tfsdk:"code_zip_b64"`
}

func (m *ServerlessDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var resource *client.ServerlessResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.ServerlessResponse](c, "serverless")
		if err != nil {
			diags.AddError(
				"Error Reading Serverless Resources",
				"Could not list serverless resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.ServerlessResponse](c, m.UUID.ValueString(), "serverless")
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading Serverless Resource"+string(code),
				"Could not read serverless resource: "+err.Error(),
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
				"Error Reading Compute Resource Status",
				"Could not read compute resource status: "+err.Error(),
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
	m.Runtime = types.StringValue(resource.Runtime)
	m.Handler = types.StringValue(resource.Handler)
	m.Memory = types.Int64Value(resource.Memory)
	m.Timeout = types.Int64Value(resource.Timeout)
	m.IsPublic = types.BoolValue(resource.IsPublic)
	m.URL = types.StringValue(resource.URL)
	m.ARN = types.StringValue(resource.ARN)
	m.State = types.StringValue(resource.State)

	credentials, err := c.GetServerlessCredentials(resource.ServiceUUID)
	if err != nil {
		diags.AddError(
			"Error Reading Serverless Credentials",
			"Could not read serverless credentials: "+err.Error(),
		)
		return true, diags
	}

	m.AccessKeyID = types.StringValue(credentials.AccessKeyID)
	m.SecretAccessKey = types.StringValue(credentials.SecretAccessKey)
	m.AuthType = types.StringValue(credentials.AuthType)

	return true, diags
}
