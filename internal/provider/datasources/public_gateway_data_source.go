package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &publicGatewayDataSource{}
	_ datasource.DataSourceWithConfigure = &publicGatewayDataSource{}
)

func NewPublicGatewayDataSource() datasource.DataSource {
	return &publicGatewayDataSource{}
}

type publicGatewayDataSource struct {
	common.BaseDataSource
}

func (d *publicGatewayDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_publicgateway"
}

func (d *publicGatewayDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the gateway.",
	}
	attributes["network_uuid"] = schema.StringAttribute{
		Computed:    true,
		Description: "UUID of the private network the gateway serves.",
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the VPC the resource belongs to.",
	}
	attributes["allocation_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "Allocation ID of the gateway address.",
	}
	attributes["nat_gateway_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the NAT gateway.",
	}
	attributes["public_subnet_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the public subnet.",
	}
	attributes["egress_enabled"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether outbound traffic is allowed.",
	}
	attributes["bytes_out_month"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Outbound traffic over the last month, in bytes.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento publicgateway resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento publicgateway resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *publicGatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.PublicGatewayDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "PublicGateway"))
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
