package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &registryDataSource{}
	_ datasource.DataSourceWithConfigure = &registryDataSource{}
)

func NewRegistryDataSource() datasource.DataSource {
	return &registryDataSource{}
}

type registryDataSource struct {
	common.BaseDataSource
}

func (d *registryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (d *registryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the registry.",
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Storage size in GB.",
	}
	attributes["is_public"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the registry is reachable from the public internet.",
	}
	attributes["repository_uri"] = schema.StringAttribute{
		Computed:    true,
		Description: "URI of the repository.",
	}
	attributes["registry_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the registry.",
	}
	attributes["image_count"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Number of images stored.",
	}
	attributes["used_bytes"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Storage in use, in bytes.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "Access key ID for the registry.",
		Sensitive:   true,
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "Secret access key for the resource.",
		Sensitive:   true,
	}
	attributes["registry_type"] = schema.StringAttribute{
		Computed:    true,
		Description: "Registry type.",
	}
	attributes["login_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "Command that logs in to the registry.",
		Sensitive:   true,
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento registry resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento registry resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *registryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.RegistryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "Registry"))
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
