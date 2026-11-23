package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &dbaasDataSource{}
	_ datasource.DataSourceWithConfigure = &dbaasDataSource{}
)

func NewDBAASDataSource() datasource.DataSource {
	return &dbaasDataSource{}
}

type dbaasDataSource struct {
	common.BaseDataSource
}

func (d *dbaasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaas"
}

func (d *dbaasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the database service.",
	}
	attributes["engine"] = schema.StringAttribute{
		Computed:    true,
		Description: "Database engine, e.g. `postgres` or `mysql`.",
	}
	attributes["backup_time"] = schema.StringAttribute{
		Computed:    true,
		Description: "Daily backup time, in `HH:MM:SS` format.",
	}
	attributes["disk_size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Disk size in GB.",
	}
	attributes["nodes_number"] = schema.Int32Attribute{
		Computed:    true,
		Description: "Number of database nodes.",
	}
	attributes["creation_date"] = schema.StringAttribute{
		Computed:    true,
		Description: "Creation timestamp.",
	}
	attributes["database_host"] = schema.StringAttribute{
		Computed:    true,
		Description: "Host to connect to.",
	}
	attributes["database_port"] = schema.StringAttribute{
		Computed:    true,
		Description: "Port to connect to.",
	}
	attributes["database_user"] = schema.StringAttribute{
		Computed:    true,
		Description: "Username for connections.",
	}
	attributes["database_password"] = schema.StringAttribute{
		Computed:  true,
		Sensitive: true,
		Description: "Password for connections.",
	}
	attributes["database_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the database to connect to.",
	}
	attributes["database_ssl_mode"] = schema.StringAttribute{
		Computed:    true,
		Description: "SSL mode for connections.",
	}
	attributes["public_uri"] = schema.StringAttribute{
		Computed:    true,
		Description: "Public connection URI.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento managed database on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento managed database on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *dbaasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.DBAASDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "DBAAS"))
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
