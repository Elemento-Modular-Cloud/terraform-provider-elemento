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
		Description: "Name of the function.",
	}
	attributes["runtime"] = schema.StringAttribute{
		Computed:    true,
		Description: "Function runtime.",
	}
	attributes["handler"] = schema.StringAttribute{
		Computed:    true,
		Description: "Entry point of the function.",
	}
	attributes["memory"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Memory allocated to the function.",
	}
	attributes["timeout"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Execution timeout of the function.",
	}
	attributes["is_public"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the function URL is reachable from the public internet.",
	}
	attributes["url"] = schema.StringAttribute{
		Computed:    true,
		Description: "Invocation URL of the function.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "ARN of the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "Access key ID for the resource.",
		Sensitive:   true,
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "Secret access key for the resource.",
		Sensitive:   true,
	}
	attributes["auth_type"] = schema.StringAttribute{
		Computed:    true,
		Description: "Authentication type of the function URL.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento serverless resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento serverless resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *ServerlessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.ServerlessDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "Serverless"))
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
