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
		Description: "Name of the public IP.",
	}
	attributes["public_ip"] = schema.StringAttribute{
		Computed:    true,
		Description: "The IP address.",
	}
	attributes["allocation_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "Allocation ID of the address.",
	}
	attributes["domain"] = schema.StringAttribute{
		Computed:    true,
		Description: "Domain of the address.",
	}
	attributes["association_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the association with an instance.",
	}
	attributes["associated_instance_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the instance the address is attached to.",
	}
	attributes["associated_network_interface_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the network interface the address is attached to.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento publicip resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento publicip resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *publicIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.PublicIPDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "PublicIP"))
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
