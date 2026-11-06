package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &ServerlessDataSource{}
	_ datasource.DataSourceWithConfigure = &ServerlessDataSource{}
)

func NewServerlessDataSource() datasource.DataSource {
	return &ServerlessDataSource{}
}

type ServerlessDataSource struct {
	common.BaseDataSource
}

func (d *ServerlessDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serverless"
}

func (d *ServerlessDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the resource.",
	}
	attributes["runtime"] = schema.StringAttribute{
		Computed:    true,
		Description: "The runtime to use for the serverless resource.",
	}
	attributes["handler"] = schema.StringAttribute{
		Computed:    true,
		Description: "The handler to use for the serverless resource.",
	}
	attributes["memory"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The memory to allocate for the serverless resource.",
	}
	attributes["timeout"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The timeout for the serverless resource.",
	}
	attributes["is_public"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the resource is public or not.",
	}
	attributes["url"] = schema.StringAttribute{
		Computed:    true,
		Description: "The URL of the resource.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ARN of the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The access key ID for the resource.",
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "The secret access key of the resource.",
	}
	attributes["auth_type"] = schema.StringAttribute{
		Computed:    true,
		Description: "The type of auth used by the resource.",
	}


	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento serverless resource on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento serverless resource on the cloud selected by `host`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *ServerlessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.ServerlessDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := model.Read(ctx, d.Client, common.AddNotFoundError(resp.Diagnostics, "Serverless"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
