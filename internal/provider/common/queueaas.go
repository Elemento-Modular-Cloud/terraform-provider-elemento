package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type QueueAASDataSourceModel struct {
	BaseDataSourceModel
	Name             types.String `tfsdk:"name"`
	NodeNumber       types.Int64  `tfsdk:"node_number"`
	NodeSize         types.String `tfsdk:"node_size"`
	SizeGB           types.Int64  `tfsdk:"size_gb"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	KafkaVersion     types.String `tfsdk:"kafka_version"`
	IsHA             types.Bool   `tfsdk:"is_ha"`
	IsPublic         types.Bool   `tfsdk:"is_public"`
	NetworkUUID      types.String `tfsdk:"network_uuid"`
	BootstrapServers types.String `tfsdk:"bootstrap_servers"`
	KafkaCert        types.String `tfsdk:"kafka_cert"`
}

type QueueAASResourceModel struct {
	QueueAASDataSourceModel
	ResourceExtraModel
}

func (m *QueueAASDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var resource *client.QueueAASResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.QueueAASResponse](c, "queueaas")
		if err != nil {
			diags.AddError(
				"Error Reading QueueAAS Resources",
				"Could not list queueaas resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.QueueAASResponse](c, m.UUID.ValueString(), "queueaas")
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading QueueAAS Resource"+string(code),
				"Could not read queueaas resource: "+err.Error(),
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
	m.NodeNumber = types.Int64Value(resource.NodeNumber)
	m.NodeSize = types.StringValue(resource.NodeSize)
	m.SizeGB = types.Int64Value(resource.SizeGB)
	m.Username = types.StringValue(resource.Username)
	m.Password = types.StringValue(resource.Password)
	m.KafkaVersion = types.StringValue(resource.KafkaVersion)
	m.IsHA = types.BoolValue(resource.IsHA)
	m.IsPublic = types.BoolValue(resource.IsPublic)
	m.NetworkUUID = types.StringValue(resource.NetworkUUID)

	credentials, err := c.GetQueueAASCredentials(resource.ServiceUUID)
	if err != nil {
		diags.AddWarning(
			"Error Reading QueueAAS Credentials",
			"Could not read queueaas credentials: "+err.Error(),
		)
		return true, diags
	}

	m.BootstrapServers = types.StringValue(credentials.BootstrapServers)
	m.KafkaCert = types.StringValue(credentials.KafkaCert)

	return true, diags
}
