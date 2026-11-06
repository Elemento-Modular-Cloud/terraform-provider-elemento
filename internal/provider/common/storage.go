package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StorageDataSourceModel struct {
	BaseDataSourceModel
	Name      types.String `tfsdk:"name"`
	Size      types.Int64  `tfsdk:"size"`
	Private   types.Bool   `tfsdk:"private"`
	Shareable types.Bool   `tfsdk:"shareable"`
	Bootable  types.Bool   `tfsdk:"bootable"`
	Readonly  types.Bool   `tfsdk:"readonly"`
	Clonable  types.Bool   `tfsdk:"clonable"`
	Priority  types.Int32  `tfsdk:"priority"`
	Tags      types.List   `tfsdk:"tags"`
}

type StorageResourceModel struct {
	StorageDataSourceModel
	ResourceExtraModel
	Algorithm      types.String `tfsdk:"alg"`
	Format         types.String `tfsdk:"format"`
	Bus            types.String `tfsdk:"bus"`
	FS             types.String `tfsdk:"fs"`
	HardwareDevice types.String `tfsdk:"hardware_device"`
	Tags           types.List   `tfsdk:"tags"`
}

func (m *StorageDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound func(ctx context.Context)) diag.Diagnostics {
	var diags diag.Diagnostics

	var resource *client.StorageResourceResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := c.ListStorageResources(m.CreatorUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading Storage Resources",
				"Could not list storage resources: "+err.Error(),
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
		res, code, err := c.GetStorageResource(m.CreatorUUID.ValueString(), m.UUID.ValueString())
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading Storage Resource",
				"Could not read storage resource: "+err.Error(),
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

	if resource == nil && (status == nil || *status == "deleted" || *status == "error" || *status == "canceled") {
		onNotFound(ctx)
		return diags
	}

	m.Status = types.StringPointerValue(status)
	m.Name = types.StringValue(resource.Name)
	m.Private = types.BoolValue(resource.Private)
	m.Shareable = types.BoolValue(resource.Shareable)
	m.Bootable = types.BoolValue(resource.Bootable)
	m.Readonly = types.BoolValue(resource.Readonly)
	m.Clonable = types.BoolValue(resource.Clonable)
	m.Size = types.Int64Value(int64(resource.Size))
	m.Priority = types.Int32Value(int32(resource.Priority))
	m.Tags, diags = types.ListValueFrom(ctx, types.StringType, resource.Tags)
	diags.Append(diags...)

	m.Region = types.StringValue(resource.Region)
	m.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	m.UUID = types.StringValue(resource.UUID)

	return diags
}
