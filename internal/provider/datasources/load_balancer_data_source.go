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
		Description: "The name of the resource.",
	}
	attributes["type"] = schema.StringAttribute{
		Computed:    true,
		Description: "The type of the resource.",
	}
	attributes["is_public"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the resource is public or not.",
	}
	attributes["port"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The port of the resource.",
	}
	attributes["scheme"] = schema.StringAttribute{
		Computed:    true,
		Description: "The scheme of the resource.",
	}
	attributes["dns_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The DNS name of the resource.",
	}
	attributes["arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "The arn of the resource.",
	}
	attributes["target_group_arn"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ARN of the target group associated with the resource.",
	}
	attributes["security_group_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the security group associated with the resource.",
	}
	attributes["vpc_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the VPC associated with the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento loadbalancer resource on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento loadbalancer resource on the cloud selected by `host`. Set exactly one of `uuid` or `billing_uuid`.",
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
