package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &storageDataSource{}
	_ datasource.DataSourceWithConfigure = &storageDataSource{}
)

func NewStorageDataSource() datasource.DataSource {
	return &storageDataSource{}
}

type storageDataSource struct {
	common.BaseDataSource
}

func (d *storageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func (d *storageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Volume name.",
	}
	attributes["size"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Volume size in bytes.",
	}
	attributes["private"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the volume is private to its owner.",
	}
	attributes["shareable"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the volume can be attached to more than one VM.",
	}
	attributes["bootable"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the volume can be used as a boot disk.",
	}
	attributes["readonly"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the volume is attached read-only.",
	}
	attributes["clonable"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the volume can be cloned.",
	}
	attributes["priority"] = schema.Int32Attribute{
		Computed:    true,
		Description: "Boot priority.",
	}
	attributes["tags"] = schema.ListAttribute{
		Computed:    true,
		ElementType: types.StringType,
		Description: "Tags attached to the volume.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento block volume on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento block volume on the cloud selected by `target`. Prefer `volume_uuid`. Set exactly one of `volume_uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *storageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.StorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "Storage"))
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
