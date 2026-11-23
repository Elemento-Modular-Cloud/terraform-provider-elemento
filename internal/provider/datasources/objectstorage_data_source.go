package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &objectStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &objectStorageDataSource{}
)

func NewObjectStorageDataSource() datasource.DataSource {
	return &objectStorageDataSource{}
}

type objectStorageDataSource struct {
	common.BaseDataSource
}

func (d *objectStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_objectstorage"
}

func (d *objectStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the bucket.",
	}
	attributes["active_storage"] = schema.Int32Attribute{
		Computed:    true,
		Description: "Storage currently in use.",
	}
	attributes["endpoint"] = schema.StringAttribute{
		Computed:    true,
		Description: "Endpoint URL of the bucket.",
	}
	attributes["account_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Storage account name.",
	}
	attributes["container_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Container name.",
	}
	attributes["sas_token"] = schema.StringAttribute{
		Computed:  true,
		Sensitive: true, Description: "SAS token granting access to the bucket.",
	}
	attributes["azure_cli_command"] = schema.StringAttribute{
		Computed:  true,
		Sensitive: true, Description: "Azure CLI command that accesses the bucket.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento object storage bucket on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento object storage bucket on the cloud selected by `target`. Set exactly one of `uuid`, `billing_uuid`, or `name`.",
		Attributes:          attributes,
	}
}

func (d *objectStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.ObjectStorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "ObjectStorage"))
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
