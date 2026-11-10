package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type KAASDataSourceModel struct {
	BaseDataSourceModel
	Name          types.String `tfsdk:"name"`
	ClusterName   types.String `tfsdk:"cluster_name"`
	ClusterStatus types.String `tfsdk:"cluster_status"`
	Endpoint      types.String `tfsdk:"endpoint"`
	KubeConfig    types.String `tfsdk:"kube_config"`
	Version       types.String `tfsdk:"kubernetes_version"`
	NetworkCIDR   types.String `tfsdk:"network_cidr"`
}

type KAASResourceModel struct {
	KAASDataSourceModel
	ResourceExtraModel
	NodePools             []nodePoolModel `tfsdk:"node_pools"`
	NodesSubnet           nodeSubnetModel `tfsdk:"nodes_subnet"`
	UpdatePolicy          types.String    `tfsdk:"update_policy"`
	KubernetesVersion     types.String    `tfsdk:"kubernetes_version"`
	ControlPlaneHAEnabled types.Bool      `tfsdk:"control_plane_ha_enabled"`
}

type nodePoolModel struct {
	Name            types.String       `tfsdk:"name"`
	AutoScale       types.Bool         `tfsdk:"auto_scale"`
	MinNodes        types.Int32        `tfsdk:"min_nodes"`
	MaxNodes        types.Int32        `tfsdk:"max_nodes"`
	SharedStorageGB types.Int64        `tfsdk:"shared_storage_gb"`
	Specs           nodePoolSpecsModel `tfsdk:"specs"`
}

type nodePoolSpecsModel struct {
	CPUCores   types.Int32 `tfsdk:"cpu_cores"`
	MemoryGB   types.Int32 `tfsdk:"memory_gb"`
	DiskSizeGB types.Int32 `tfsdk:"disk_size_gb"`
	GPUs       types.Int32 `tfsdk:"gpus"`
}

type nodeSubnetModel struct {
	DHCPEnabled types.Bool   `tfsdk:"dhcp_enabled"`
	CIDR        types.String `tfsdk:"cidr"`
}

func (m *KAASDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) diag.Diagnostics {
	var diags diag.Diagnostics
	var resource *client.KAASResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := c.ListKAASResources()
		if err != nil {
			diags.AddError(
				"Error Reading KAAS Resources",
				"Could not list kaas resources: "+err.Error(),
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
		res, code, err := c.GetKAASResource(m.UUID.ValueString())
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading KAAS Resource",
				"Could not read kaas resource: "+err.Error(),
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
	m.UUID = types.StringValue(resource.UUID)
	m.CreatorUUID = types.StringValue(resource.CreatorUUID)

	m.ClusterName = types.StringValue(resource.ClusterName)
	m.ClusterStatus = types.StringValue(resource.Status)
	m.Endpoint = types.StringValue(resource.Endpoint)
	m.KubeConfig = types.StringValue(resource.KubeConfig)
	m.NetworkCIDR = types.StringValue(resource.NetworkCIDR)

	// credentials, err := c.GetKAASCredentials(resource.UUID)
	// if err != nil {
	//     diags.AddError(
	//         "Error Reading KAAS Credentials",
	//         "Could not read kaas credentials: "+err.Error(),
	//     )
	//     return
	// }

	return diags
}
