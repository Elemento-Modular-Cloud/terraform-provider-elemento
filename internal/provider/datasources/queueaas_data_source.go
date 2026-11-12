package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &queueAASDataSource{}
	_ datasource.DataSourceWithConfigure = &queueAASDataSource{}
)

func NewQueueAASDataSource() datasource.DataSource {
	return &queueAASDataSource{}
}

type queueAASDataSource struct {
	common.BaseDataSource
}

func (d *queueAASDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_queueaas"
}

func (d *queueAASDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the resource.",
	}
	attributes["node_number"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The number of nodes of the queue.",
	}
	attributes["node_size"] = schema.StringAttribute{
		Computed:    true,
		Description: "The size of each node of the queue.",
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The size (in GB) of the queue.",
	}
	attributes["username"] = schema.StringAttribute{
		Computed:    true,
		Description: "The username to use for the queue.",
	}
	attributes["password"] = schema.StringAttribute{
		Computed:    true,
		Description: "The password to use for the queue.",
		Sensitive:   true,
	}
	attributes["kafka_version"] = schema.StringAttribute{
		Computed:    true,
		Description: "The version of Kafka to use for the queue.",
	}
	attributes["is_ha"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether HA should be enabled for the resource or not.",
	}
	attributes["is_public"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whetherthe resource should be public or not.",
	}
	attributes["network_uuid"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the model used by the resource.",
	}
	attributes["bootstrap_servers"] = schema.StringAttribute{
		Computed:    true,
		Description: "The boostrap servers of the resource.",
	}
	attributes["kafka_cert"] = schema.StringAttribute{
		Computed:    true,
		Description: "The Kafka certificate for the resource.",
		Sensitive:   true,
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento queueaas resource on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento queueaas resource on the cloud selected by `host`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *queueAASDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.QueueAASDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "QueueAAS"))
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
