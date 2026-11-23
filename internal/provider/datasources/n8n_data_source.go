package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &n8nDataSource{}
	_ datasource.DataSourceWithConfigure = &n8nDataSource{}
)

func NewN8NDataSource() datasource.DataSource {
	return &n8nDataSource{}
}

type n8nDataSource struct {
	common.BaseDataSource
}

func (d *n8nDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_n8n"
}

func (d *n8nDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["vm_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the virtual machine running n8n.",
	}
	attributes["cpu_cores"] = schema.Int32Attribute{
		Computed:    true,
		Description: "Number of CPU cores.",
	}
	attributes["ram_size_mb"] = schema.Int32Attribute{
		Computed:    true,
		Description: "RAM in MB.",
	}
	attributes["disk_size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Boot disk size in GB.",
	}
	attributes["service_url"] = schema.StringAttribute{
		Computed:    true,
		Description: "URL of the n8n instance.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Deployment state, e.g. `ERROR` or `SUCCESS`.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento n8n resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento n8n resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *n8nDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.N8NDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "N8N"))
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
