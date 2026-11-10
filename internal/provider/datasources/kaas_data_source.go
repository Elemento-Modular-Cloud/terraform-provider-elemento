package datasources

import (
	"context"

	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &kaasDataSource{}
	_ datasource.DataSourceWithConfigure = &kaasDataSource{}
)

func NewKAASDataSource() datasource.DataSource {
	return &kaasDataSource{}
}

type kaasDataSource struct {
	common.BaseDataSource
}

func (d *kaasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kaas"
}

func (d *kaasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := common.BaseDataSourceSchema()

	attributes["name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Cluster name as returned by the API, or the Terraform-facing name. Not unique.",
	}
	attributes["cluster_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "Kubernetes cluster name.",
	}
	attributes["cluster_status"] = schema.StringAttribute{
		Computed:    true,
		Description: "Cluster status from Meson (for example PENDING or RUNNING).",
	}
	attributes["endpoint"] = schema.StringAttribute{
		Computed:    true,
		Description: "Kubernetes API endpoint.",
	}
	attributes["kube_config"] = schema.StringAttribute{
		Computed:    true,
		Sensitive:   true,
		Description: "Kubeconfig. Empty while the cluster is still pending.",
	}
	attributes["kubernetes_version"] = schema.StringAttribute{
		Computed:    true,
		Description: "Kubernetes version.",
	}
	attributes["network_cidr"] = schema.StringAttribute{
		Computed:    true,
		Description: "Cluster network CIDR.",
	}

	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento managed Kubernetes cluster on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento managed Kubernetes cluster on the cloud selected by `host`. Set exactly one of `uuid`, `billing_uuid`, or `name`. A `PENDING` cluster may have an empty kubeconfig.",
		Attributes:          attributes,
	}
}

func (d *kaasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model common.KAASDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, d.Client, common.AddNotFoundError(resp.Diagnostics, "KAAS"))
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
