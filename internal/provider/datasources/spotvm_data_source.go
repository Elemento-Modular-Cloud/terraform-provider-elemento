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
			Description: "Volume name.",
		},
		"size_gb": schema.Int64Attribute{
			Computed:    true,
			Description: "Volume size in GB.",
		},
	}

	attributes["vm_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the VM.",
	}

	attributes["cpu"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "CPU requirements.",
		Attributes: map[string]schema.Attribute{
			"slots": schema.Int32Attribute{
				Description: "Number of CPU slots.",
				Computed:    true,
			},
			"shared_cores": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether shared cores are allowed.",
			},
			"arch": schema.ListAttribute{
				Computed:    true,
				Description: "Accepted CPU architectures: `X86_64`, `arm64`.",
				ElementType: types.StringType,
			},
		},
	}

	attributes["mem"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "Memory requirements.",
		Attributes: map[string]schema.Attribute{
			"capacity_mb": schema.Int32Attribute{
				Computed:    true,
				Description: "Memory in MB.",
			},
		},
	}

	attributes["os"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "Operating system requirements.",
		Attributes: map[string]schema.Attribute{
			"family": schema.StringAttribute{
				Description: "OS family: `linux`, `windows` or `macos`.",
				Computed:    true,
			},
			"flavour": schema.StringAttribute{
				Description: "OS flavour.",
				Computed:    true, // TODO: os_flavour validator based on iso-templates?
			},
			"version": schema.StringAttribute{
				Description: "OS version.",
				Computed:    true,
			},
		},
	}

	attributes["storage"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "Storage requirements.",
		Attributes: map[string]schema.Attribute{
			"boot": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Boot volume.",
				Attributes:  volumeAttributes,
			},
			"data": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Additional data volumes.",
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
		Description: "Network interfaces of the VM.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"kind": schema.StringAttribute{
					Computed:    true,
					Description: "Network kind.",
				},
				"ipv4": schema.StringAttribute{
					Computed:    true,
					Description: "IPv4 address.",
				},
				"mac": schema.StringAttribute{
					Computed:    true,
					Description: "MAC address.",
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
		Description: "Tags attached to the VM.",
		ElementType: types.StringType,
	}

	attributes["cloudinit_b64"] = schema.StringAttribute{
		Computed:    true,
		Description: "Base64-encoded cloud-init file.",
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

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "SpotVM"))
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
