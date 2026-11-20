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
		Description: "The name of the resource.",
	}
	attributes["network_uuid"] = schema.StringAttribute{
		Computed:    true,
		Description: "The UUID of the private network associated with the resource.",
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the VPC associated with the resource.",
	}
	attributes["allocation_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The allocation ID of the resource.",
	}
	attributes["nat_gateway_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the nat gateway of the resource.",
	}
	attributes["public_subnet_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the public subnet associated with the resource.",
	}
	attributes["egress_enabled"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether egress is enabled for this public gateway.",
	}
	attributes["bytes_out_month"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The traffic (in bytes) sustained last month by the public gateway.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
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

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "PublicGateway"))
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
