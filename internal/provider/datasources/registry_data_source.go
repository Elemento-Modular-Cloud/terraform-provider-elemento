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
		Description: "The name of the resource.",
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The size (in GB) of the resource.",
	}
	attributes["is_public"] = schema.BoolAttribute{
		Computed:    true,
		Description: "The size (in GB) of the resource.",
	}
	attributes["repository_uri"] = schema.StringAttribute{
		Computed:    true,
		Description: "The URI of the repository.",
	}
	attributes["registry_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the registry.",
	}
	attributes["image_count"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The number of images in the registry.",
	}
	attributes["used_bytes"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The size (in bytes) currently used by the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the access key of the resource.",
		Sensitive:   true,
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "The secret access key of the resource.",
		Sensitive:   true,
	}
	attributes["registry_type"] = schema.StringAttribute{
		Computed:    true,
		Description: "The type of the registry.",
	}
	attributes["login_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "The command used to login to the resource.",
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

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "Registry"))
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
