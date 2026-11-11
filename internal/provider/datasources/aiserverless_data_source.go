package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &aiServerlessDataSource{}
	_ datasource.DataSourceWithConfigure = &aiServerlessDataSource{}
)

func NewAIServerlessDataSource() datasource.DataSource {
	return &aiServerlessDataSource{}
}

type aiServerlessDataSource struct {
	common.BaseDataSource
}

func (d *aiServerlessDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aiserverless"
}

func (d *aiServerlessDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the AIServerless resource.",
	}
	attributes["model_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the model to use for the aiserverless resource.",
	}
	attributes["enable_streaming"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether streaming should be enabled for the resource or not.",
	}
	attributes["model_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the model used by the resource.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ARN of the resource.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The access key ID for the resource.",
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "The secret access key of the resource.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento aiserverless resource on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento aiserverless resource on the cloud selected by `host`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *aiServerlessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.AIServerlessDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "AIServerless"))
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
