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
		Description: "Name of the file storage.",
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Provisioned size in GB.",
	}
	attributes["used_gb"] = schema.Float64Attribute{
		Computed:    true,
		Description: "Space currently used, in GB.",
	}
	attributes["dns_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "DNS name of the NFS server.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
	}
	attributes["nfs_host"] = schema.StringAttribute{
		Computed:    true,
		Description: "NFS server host.",
	}
	attributes["nfs_path"] = schema.StringAttribute{
		Computed:    true,
		Description: "Exported NFS path.",
	}
	attributes["nfs_version"] = schema.StringAttribute{
		Computed:    true,
		Description: "NFS protocol version.",
	}
	attributes["mount_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "Command that mounts the share.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento filestorage on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento filestorage on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
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
