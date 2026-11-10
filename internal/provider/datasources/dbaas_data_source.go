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
		Description: "DB name.",
	}
	attributes["engine"] = schema.StringAttribute{
		Computed:    true,
		Description: "Database engine (for example postgres or mysql).",
	}
	attributes["backup_time"] = schema.StringAttribute{
		Computed:    true,
		Description: "Daily backup time.",
	}
	attributes["disk_size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Disk size as returned by the API.",
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
		Description: "Database host.",
	}
	attributes["database_port"] = schema.StringAttribute{
		Computed:    true,
		Description: "Database port.",
	}
	attributes["database_user"] = schema.StringAttribute{
		Computed:    true,
		Description: "Database username.",
	}
	attributes["database_password"] = schema.StringAttribute{
		Computed:  true,
		Sensitive: true, Description: "Database password.",
	}
	attributes["database_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Database name used for connections.",
	}
	attributes["database_ssl_mode"] = schema.StringAttribute{
		Computed:    true,
		Description: "SSL mode for the database connection.",
	}
	attributes["public_uri"] = schema.StringAttribute{
		Computed:    true,
		Description: "Public connection URI.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento managed database on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento managed database on the cloud selected by `host`. Set exactly one of `uuid`, `billing_uuid`, or `name`. Names are not unique; multiple matches return an error.",
		Attributes:          attributes,
	}
}

func (d *dbaasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.DBAASDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(resp.Diagnostics, "DBAAS"))
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
