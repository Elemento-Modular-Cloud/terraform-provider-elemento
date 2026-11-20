package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &openClawDataSource{}
	_ datasource.DataSourceWithConfigure = &openClawDataSource{}
)

func NewOpenClawDataSource() datasource.DataSource {
	return &openClawDataSource{}
}

type openClawDataSource struct {
	common.BaseDataSource
}

func (d *openClawDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openclaw"
}

func (d *openClawDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["vm_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the virtual machine running OpenClaw.",
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
		Description: "URL of the OpenClaw instance.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Deployment state, e.g. `ERROR` or `SUCCESS`.",
	}

	attributes["token"] = schema.StringAttribute{
		Computed:    true,
		Sensitive:   true,
		Description: "Access token for the instance.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento openclaw resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento openclaw resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *openClawDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.OpenClawDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "OpenClaw"))
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
