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
		Description: "The name of the virtual machine for the n8n resource.",
	}
	attributes["cpu_cores"] = schema.Int32Attribute{
		Computed:    true,
		Description: "The number of CPU cores for the n8n resource.",
	}
	attributes["ram_size_mb"] = schema.Int32Attribute{
		Computed:    true,
		Description: "The amount of RAM in MB for the n8n resource.",
	}
	attributes["disk_size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The size (in GB) of the boot disk for the n8n resource.",
	}
	attributes["service_url"] = schema.StringAttribute{
		Computed:    true,
		Description: "The service URL for the n8n resource, assigned after creation.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the n8n resource (e.g., 'ERROR', 'SUCCESS').",
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

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "N8N"))
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
