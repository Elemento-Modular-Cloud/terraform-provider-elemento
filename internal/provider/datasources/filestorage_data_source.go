package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &fileStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &fileStorageDataSource{}
)

func NewFileStorageDataSource() datasource.DataSource {
	return &fileStorageDataSource{}
}

type fileStorageDataSource struct {
	common.BaseDataSource
}

func (d *fileStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filestorage"
}

func (d *fileStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the resource.",
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The size (in GB) of the resource.",
	}
	attributes["used_gb"] = schema.Float64Attribute{
		Computed:    true,
		Description: "The size (in GB) currently used by the resource.",
	}
	attributes["dns_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the DNS associated with the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}
	attributes["nfs_host"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs host associated with the resource",
	}
	attributes["nfs_path"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs path associated with the resource",
	}
	attributes["nfs_version"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs version associated with the resource",
	}
	attributes["mount_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "The command used to connect to the resource",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento filestorage on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento filestorage on the cloud selected by `host`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *fileStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.FileStorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "FileStorage"))
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
