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
		Description: "The name of the virtual machine for the openclaw resource.",
	}
	attributes["cpu_cores"] = schema.Int32Attribute{
		Computed:    true,
		Description: "The number of CPU cores for the openclaw resource.",
	}
	attributes["ram_size_mb"] = schema.Int32Attribute{
		Computed:    true,
		Description: "The amount of RAM in MB for the openclaw resource.",
	}
	attributes["disk_size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The size (in GB) of the boot disk for the openclaw resource.",
	}
	attributes["service_url"] = schema.StringAttribute{
		Computed:    true,
		Description: "The service URL for the openclaw resource, assigned after creation.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the openclaw resource (e.g., 'ERROR', 'SUCCESS').",
	}

	attributes["token"] = schema.StringAttribute{
		Computed:    true,
		Sensitive:   true,
		Description: "The token for the openclaw resource, assigned after creation.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento openclaw resource on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento openclaw resource on the cloud selected by `host`. Set exactly one of `uuid` or `billing_uuid`.",
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
