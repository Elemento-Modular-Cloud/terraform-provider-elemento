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
	VMName         types.String       `tfsdk:"vm_name"`
	CPU            spotVMCPUModel     `tfsdk:"cpu"`
	Mem            spotVMMemoryModel  `tfsdk:"mem"`
	OS             spotVMOSModel      `tfsdk:"os"`
	Storage        spotVMStorageModel `tfsdk:"storage"`
	Networks       types.List         `tfsdk:"networks"`
	Authentication *spotVMAuthModel   `tfsdk:"auth"`
	CloudInitB64   types.String       `tfsdk:"cloudinit_b64"`
	Tags           types.List         `tfsdk:"tags"`

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

func (m *SpotVMDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound func(ctx context.Context)) diag.Diagnostics {
	var diags diag.Diagnostics

	resources, err := c.ListSpotVMResources()
	if err != nil {
		diags.AddError(
			"Error Reading SpotVM Resources",
			"Could not list spotvm resources: "+err.Error(),
		)
		return diags
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
		status, err = c.GetResourceStatus(m.BillingUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading SpotVM Resource Status",
				"Could not read spotvm resource status: "+err.Error(),
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

	m.CPU.Arch, diags = types.ListValueFrom(ctx, types.StringType, resource.CPU.Arch)
	diags.Append(diags...)

	m.CPU.Slots = types.Int32Value(int32(resource.CPU.Slots))
	m.CPU.SharedCores = types.BoolValue(resource.CPU.SharedCores)

	m.Mem.CapacityMB = types.Int32Value(int32(resource.Mem.CapacityMB))

	// TODO: PCI DEVICES

	m.OS.Family = types.StringValue(resource.OS.Family)
	m.OS.Flavour = types.StringValue(resource.OS.Flavour)
	m.OS.Version = types.StringValue(resource.OS.Version)

	var dataVolumes []spotVMVolumeModel
	for _, volume := range resource.Storage.Data {
		dataVolumes = append(dataVolumes, spotVMVolumeModel{
			Name:   types.StringValue(volume.Name),
			SizeGB: types.Int64Value(int64(volume.SizeGB)),
		})
	}
	m.Storage.Data = dataVolumes

	var networks []spotVMNetworkModel
	for _, network := range resource.Networks {
		networks = append(networks, spotVMNetworkModel{
			Kind: types.StringValue(network.Kind),
			IPV4: types.StringPointerValue(network.IPv4),
			MAC:  types.StringPointerValue(network.MAC),
		})
	}
	m.Networks, diags = types.ListValueFrom(ctx, SpotVMNetworkObjectType(), networks)

	m.Status = types.StringPointerValue(status)
	m.VMName = types.StringValue(resource.VMName)
	m.Region = types.StringValue(resource.Region)
	m.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	m.UUID = types.StringValue(resource.UUID)
	m.Tags, diags = types.ListValueFrom(ctx, types.StringType, resource.Tags)
	diags.Append(diags...)

	return diags
}
