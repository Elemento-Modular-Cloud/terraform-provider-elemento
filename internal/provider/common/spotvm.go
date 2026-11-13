package common

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SpotVMDataSourceModel struct {
	BaseDataSourceModel
	VMName         types.String        `tfsdk:"vm_name"`
	CPU            *spotVMCPUModel     `tfsdk:"cpu"`
	Mem            *spotVMMemoryModel  `tfsdk:"mem"`
	OS             *spotVMOSModel      `tfsdk:"os"`
	Storage        *spotVMStorageModel `tfsdk:"storage"`
	Networks       types.List          `tfsdk:"networks"`
	Authentication *spotVMAuthModel    `tfsdk:"auth"`
	CloudInitB64   types.String        `tfsdk:"cloudinit_b64"`
	Tags           types.List          `tfsdk:"tags"`
}

type SpotVMResourceModel struct {
	SpotVMDataSourceModel
	ResourceExtraModel
}

type spotVMCPUModel struct {
	Slots       types.Int32 `tfsdk:"slots"`
	Arch        types.List  `tfsdk:"arch"`
	SharedCores types.Bool  `tfsdk:"shared_cores"`
}

type spotVMMemoryModel struct {
	CapacityMB types.Int32 `tfsdk:"capacity_mb"`
}

type spotVMOSModel struct {
	Family  types.String `tfsdk:"family"`
	Flavour types.String `tfsdk:"flavour"`
	Version types.String `tfsdk:"version"`
}

type spotVMStorageModel struct {
	Boot spotVMVolumeModel   `tfsdk:"boot"`
	Data []spotVMVolumeModel `tfsdk:"data"`
}

type spotVMNetworkModel struct {
	Kind types.String `tfsdk:"kind"`
	IPV4 types.String `tfsdk:"ipv4"`
	MAC  types.String `tfsdk:"mac"`
}

type spotVMVolumeModel struct {
	Name   types.String `tfsdk:"name"`
	SizeGB types.Int64  `tfsdk:"size_gb"`
}

type spotVMAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	SSHKey   types.String `tfsdk:"ssh_key"`
}

func SpotVMNetworkObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"kind": types.StringType,
			"ipv4": types.StringType,
			"mac":  types.StringType,
		},
	}
}

func (m *SpotVMDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	resources, err := c.ListSpotVMResources(ctx)
	if err != nil {
		diags.AddError(
			"Error Reading SpotVM Resources",
			"Could not list spotvm resources: "+err.Error(),
		)
		return true, diags
	}

	var resource *client.SpotVMListItem
	for _, res := range resources {
		if res.BillingUUID == m.BillingUUID.ValueString() || res.UUID == m.UUID.ValueString() {
			resource = &res
			break
		}
	}

	billing_uuid := m.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		status, err = c.GetResourceStatus(ctx, m.BillingUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading SpotVM Resource Status",
				"Could not read spotvm resource status: "+err.Error(),
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

	arch, tmpDiags := types.ListValueFrom(ctx, types.StringType, resource.CPU.Arch)
	diags.Append(tmpDiags...)

	m.CPU = &spotVMCPUModel{
		Slots: types.Int32Value(int32(resource.CPU.Slots)),
		SharedCores: types.BoolValue(resource.CPU.SharedCores),
		Arch: arch,
	}

	m.Mem = &spotVMMemoryModel{
		CapacityMB: types.Int32Value(int32(resource.Mem.CapacityMB)),
	}

	// TODO: PCI DEVICES

	m.OS = &spotVMOSModel{
		Family: types.StringValue(resource.OS.Family),
		Flavour: types.StringValue(resource.OS.Flavour),
		Version: types.StringValue(resource.OS.Version),
	}

	var dataVolumes []spotVMVolumeModel
	for _, volume := range resource.Storage.Data {
		dataVolumes = append(dataVolumes, spotVMVolumeModel{
			Name:   types.StringValue(volume.Name),
			SizeGB: types.Int64Value(int64(volume.SizeGB)),
		})
	}

	m.Storage = &spotVMStorageModel{
		Data: dataVolumes,
		Boot: spotVMVolumeModel{
			Name:   types.StringValue(resource.Storage.Boot.Name),
			SizeGB: types.Int64Value(int64(resource.Storage.Boot.SizeGB)),
		},
	}

	var networks []spotVMNetworkModel
	for _, network := range resource.Networks {
		networks = append(networks, spotVMNetworkModel{
			Kind: types.StringValue(network.Kind),
			IPV4: types.StringPointerValue(network.IPv4),
			MAC:  types.StringPointerValue(network.MAC),
		})
	}
	m.Networks, tmpDiags = types.ListValueFrom(ctx, SpotVMNetworkObjectType(), networks)
	diags.Append(tmpDiags...)

	m.Status = types.StringPointerValue(status)
	m.VMName = types.StringValue(resource.VMName)
	m.Region = types.StringValue(resource.Region)
	m.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	m.UUID = types.StringValue(resource.UUID)
	m.Tags, tmpDiags = types.ListValueFrom(ctx, types.StringType, resource.Tags)
	diags.Append(tmpDiags...)

	return true, diags
}
