package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &spotVMDataSource{}
	_ datasource.DataSourceWithConfigure = &spotVMDataSource{}
)

func NewSpotVMDataSource() datasource.DataSource {
	return &spotVMDataSource{}
}

type spotVMDataSource struct {
	common.BaseDataSource
}

func (d *spotVMDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spotvm"
}

func (d *spotVMDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	volumeAttributes := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "The name of the volume.",
		},
		"size_gb": schema.Int64Attribute{
			Computed:    true,
			Description: "The size of the volume, in GB.",
		},
	}

	attributes["vm_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the spotvm resource.",
	}

	attributes["cpu"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "The elemento spotvm cpu requirements",
		Attributes: map[string]schema.Attribute{
			"slots": schema.Int32Attribute{
				Description: "The number of CPU slots Computed for the spotvm resource.",
				Computed:    true,
			},
			"shared_cores": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether shared cores are allowed for the CPU.",
			},
			"arch": schema.ListAttribute{
				Computed:    true,
				Description: "The list of alternative CPU architectures Computed for the spotvm resource. Valid values are 'X86_64' and 'arm64'.",
				ElementType: types.StringType,
			},
		},
	}

	attributes["mem"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "The elemento spotvm memory requirements",
		Attributes: map[string]schema.Attribute{
			"capacity_mb": schema.Int32Attribute{
				Computed:    true,
				Description: "The memory capacity (in MB) Computed for the spotvm resource.",
			},
		},
	}

	attributes["os"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "The elemento spotvm miscellaneous requirements",
		Attributes: map[string]schema.Attribute{
			"family": schema.StringAttribute{
				Description: "The operating system family Computed for the spotvm resource. Valid values are 'linux', 'windows', and 'macos'.",
				Computed:    true,
			},
			"flavour": schema.StringAttribute{
				Description: "The operating system flavour Computed for the spotvm resource.",
				Computed:    true, // TODO: os_flavour validator based on iso-templates?
			},
			"version": schema.StringAttribute{
				Description: "The operating system version Computed for the spotvm resource.",
				Computed:    true,
			},
		},
	}

	attributes["storage"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "The storage Computed for the spotvm resource.",
		Attributes: map[string]schema.Attribute{
			"boot": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The boot volume Computed for the spotvm storage.",
				Attributes:  volumeAttributes,
			},
			"data": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of Computed data volumes Computed for the spotvm resource.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: volumeAttributes,
				},
			},
		},
	}

	// // TODO: Handle PCI
	// attributes["pci"] = schema.SingleNestedAttribute{
	// 	Computed:    true,
	// 	Description: "The elemento spotvm PCI requirements",
	// 	Attributes: map[string]schema.Attribute{
	// 		"devices": schema.ListAttribute{
	// 			Computed:    true,
	// 			Description: "A list of PCI device IDs Computed for the spotvm resource.",
	// 			ElementType: types.StringType,
	// 		},
	// 	},
	// }

	attributes["networks"] = schema.ListNestedAttribute{
		Computed:    true,
		Description: "The list of networks associated with the spotvm resource.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"kind": schema.StringAttribute{
					Computed:    true,
					Description: "The network kind.",
				},
				"ipv4": schema.StringAttribute{
					Computed:    true,
					Description: "The IPv4 address of the network.",
				},
				"mac": schema.StringAttribute{
					Computed:    true,
					Description: "The MAC address of the network.",
				},
			},
		},
	}

	attributes["auth"] = schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Computed: true,
			},
			"password": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"ssh_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}

	attributes["tags"] = schema.ListAttribute{
		Computed:    true,
		Description: "The tags associated with the spotvm resource, assigned by the Elemento API after creation.",
		ElementType: types.StringType,
	}

	attributes["cloudinit_b64"] = schema.StringAttribute{
		Computed:    true,
		Description: "The cloudinit file for the spotvm, base64-encoded.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento spot vm on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento spot vm on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *spotVMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.SpotVMDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "SpotVM"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || !found {
		return
	}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
