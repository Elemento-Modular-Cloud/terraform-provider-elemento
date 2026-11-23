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
		Description: "Name of the queue.",
	}
	attributes["node_number"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Number of nodes.",
	}
	attributes["node_size"] = schema.StringAttribute{
		Computed:    true,
		Description: "Size of each node.",
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Storage size in GB.",
	}
	attributes["username"] = schema.StringAttribute{
		Computed:    true,
		Description: "Username for connections.",
	}
	attributes["password"] = schema.StringAttribute{
		Computed:    true,
		Description: "Password for connections.",
		Sensitive:   true,
	}
	attributes["kafka_version"] = schema.StringAttribute{
		Computed:    true,
		Description: "Kafka version.",
	}
	attributes["is_ha"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the cluster runs in high availability.",
	}
	attributes["is_public"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the cluster is reachable from the public internet.",
	}
	attributes["network_uuid"] = schema.StringAttribute{
		Computed:    true,
		Description: "UUID of the private network to attach to.",
	}
	attributes["bootstrap_servers"] = schema.StringAttribute{
		Computed:    true,
		Description: "Kafka bootstrap servers.",
	}
	attributes["kafka_cert"] = schema.StringAttribute{
		Computed:    true,
		Description: "Kafka certificate.",
		Sensitive:   true,
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento queueaas resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento queueaas resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *queueAASDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.QueueAASDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "QueueAAS"))
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
