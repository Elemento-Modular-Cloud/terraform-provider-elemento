package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &loadBalancerDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerDataSource{}
)

func NewLoadBalancerDataSource() datasource.DataSource {
	return &loadBalancerDataSource{}
}

type loadBalancerDataSource struct {
	common.BaseDataSource
}

func (d *loadBalancerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

func (d *loadBalancerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Name of the load balancer.",
	}
	attributes["type"] = schema.StringAttribute{
		Computed:    true,
		Description: "Load balancer type.",
	}
	attributes["is_public"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the load balancer is internet-facing.",
	}
	attributes["port"] = schema.Int64Attribute{
		Computed:    true,
		Description: "Listener port.",
	}
	attributes["scheme"] = schema.StringAttribute{
		Computed:    true,
		Description: "Listener scheme.",
	}
	attributes["dns_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "DNS name of the load balancer.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "ARN of the load balancer.",
	}
	attributes["target_group_arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "ARN of the target group.",
	}
	attributes["security_group_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the security group.",
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "ID of the VPC the resource belongs to.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Current state reported by the service.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento loadbalancer resource on the cloud selected by target.",
		MarkdownDescription: "Looks up an existing Elemento loadbalancer resource on the cloud selected by `target`. Set exactly one of `uuid` or `billing_uuid`.",
		Attributes:          attributes,
	}
}

func (d *loadBalancerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.LoadBalancerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(&resp.Diagnostics, "LoadBalancer"))
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
