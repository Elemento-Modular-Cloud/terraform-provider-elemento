package common

import (
	"context"
	"strconv"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AIServerlessDataSourceModel struct {
	BaseDataSourceModel
	Name            types.String `tfsdk:"name"`
	ModelName       types.String `tfsdk:"model_name"`
	EnableStreaming types.Bool   `tfsdk:"enable_streaming"`
	ModelID         types.String `tfsdk:"model_id"`
	ARN             types.String `tfsdk:"arn"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

type AIServerlessResourceModel struct {
	AIServerlessDataSourceModel
	ResourceExtraModel
}

func (m *AIServerlessDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var resource *client.AIServerlessResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.AIServerlessResponse](c, ctx, "aiserverless")
		if err != nil {
			diags.AddError(
				"Error Reading AIServerless Resources",
				"Could not list aiserverless resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.AIServerlessResponse](c, ctx, m.UUID.ValueString(), "aiserverless")
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading AIServerless Resource"+strconv.Itoa(code),
				"Could not read aiserverless resource: "+err.Error(),
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
				"Error Reading AIServerless Resource Status",
				"Could not read aiserverless resource status: "+err.Error(),
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
	m.ModelName = types.StringValue(resource.ModelName)
	m.EnableStreaming = types.BoolValue(resource.EnableStreaming)
	m.ModelID = types.StringValue(resource.ModelID)
	m.ARN = types.StringValue(resource.ARN)

	credentials, err := c.GetAIServerlessCredentials(ctx, resource.ServiceUUID)
	if err != nil {
		diags.AddWarning(
			"Error Reading AIServerless Credentials",
			"Could not read aiserverless credentials: "+err.Error(),
		)
		return true, diags
	}

	m.AccessKeyID = types.StringValue(credentials.AccessKeyID)
	m.SecretAccessKey = types.StringValue(credentials.SecretAccessKey)

	return true, diags
}
