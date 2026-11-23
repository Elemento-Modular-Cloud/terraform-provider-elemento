package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &privateNetworkDataSource{}
	_ datasource.DataSourceWithConfigure = &privateNetworkDataSource{}
)

func NewPrivateNetworkDataSource() datasource.DataSource {
	return &privateNetworkDataSource{}
}

type privateNetworkDataSource struct {
	common.BaseDataSource
}

func (d *privateNetworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privatenetwork"
}

func (d *privateNetworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the network.",
	}
	attributes["cidr"] = schema.StringAttribute{
		Computed:    true,
		Description: "CIDR block of the network.",
	}
	attributes["auto_assign_cidr"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Let the platform pick the CIDR block.",
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the VPC the resource belongs to.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
	}
	attributes["subnets"] = schema.ListNestedAttribute{
		Computed:    true,
		Description: "Subnets of the network.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"az": schema.StringAttribute{
					Computed:    true,
					Description: "Availability zone of the subnet.",
				},
				"subnet_id": schema.StringAttribute{
					Computed:    true,
					Description: "ID of the subnet.",
				},
				"cidr": schema.StringAttribute{
					Computed:    true,
					Description: "CIDR block of the subnet.",
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento privatenetwork resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento privatenetwork resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *privateNetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.PrivateNetworkDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "PrivateNetwork"))
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
