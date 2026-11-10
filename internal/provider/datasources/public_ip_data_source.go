package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &publicIPDataSource{}
	_ datasource.DataSourceWithConfigure = &publicIPDataSource{}
)

func NewPublicIPDataSource() datasource.DataSource {
	return &publicIPDataSource{}
}

type publicIPDataSource struct {
	common.BaseDataSource
}

func (d *publicIPDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_publicip"
}

func (d *publicIPDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the resource.",
	}
	attributes["public_ip"] = schema.StringAttribute{
		Computed:    true,
		Description: "The public IP of the resource.",
	}
	attributes["allocation_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The allocation ID of the resource.",
	}
	attributes["domain"] = schema.StringAttribute{
		Computed:    true,
		Description: "The domain of the resource.",
	}
	attributes["association_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The association ID of the resource.",
	}
	attributes["associated_instance_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the instance associated with the resource.",
	}
	attributes["associated_network_interface_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the network interface associated with the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento publicip resource on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento publicip resource on the cloud selected by `host`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *publicIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.PublicIPDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(resp.Diagnostics, "PublicIP"))
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
