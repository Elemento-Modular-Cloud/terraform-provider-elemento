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
		Description: "The name of the resource.",
	}
	attributes["cidr"] = schema.StringAttribute{
		Computed:    true,
		Description: "The network CIDR of the resource.",
	}
	attributes["auto_assign_cidr"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the CIDR should be automatically assigned for this resource.",
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The VPC ID of the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}
	attributes["subnets"] = schema.ListNestedAttribute{
		Computed:    true,
		Description: "The subnets associated with the resource",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"az": schema.StringAttribute{
					Computed:    true,
					Description: "The availability zone of the subnet",
				},
				"subnet_id": schema.StringAttribute{
					Computed:    true,
					Description: "The ID of the subnet",
				},
				"cidr": schema.StringAttribute{
					Computed:    true,
					Description: "The CIDR of the subnet",
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento privatenetwork resource on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento privatenetwork resource on the cloud selected by `host`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *privateNetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.PrivateNetworkDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := model.Read(ctx, d.Client, common.AddNotFoundError(resp.Diagnostics, "PrivateNetwork"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
