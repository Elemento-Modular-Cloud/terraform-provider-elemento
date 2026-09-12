package provider

import (
	"context"

	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &kaasDataSource{}
	_ datasource.DataSourceWithConfigure = &kaasDataSource{}
)

func NewKAASDataSource() datasource.DataSource {
	return &kaasDataSource{}
}

type kaasDataSource struct {
	client *client.ElementoClient
}

type kaasDataSourceModel struct {
	UUID          types.String `tfsdk:"uuid"`
	BillingUUID   types.String `tfsdk:"billing_uuid"`
	Name          types.String `tfsdk:"name"`
	Region        types.String `tfsdk:"region"`
	CreatorUUID   types.String `tfsdk:"creator_uuid"`
	ClusterName   types.String `tfsdk:"cluster_name"`
	ClusterStatus types.String `tfsdk:"cluster_status"`
	Endpoint      types.String `tfsdk:"endpoint"`
	KubeConfig    types.String `tfsdk:"kube_config"`
	Version       types.String `tfsdk:"kubernetes_version"`
	NetworkCIDR   types.String `tfsdk:"network_cidr"`
	Status        types.String `tfsdk:"status"`
}

func (d *kaasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kaas"
}

func (d *kaasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ids := identifierExactlyOneOf("uuid", "billing_uuid", "name")
	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento managed Kubernetes cluster on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento managed Kubernetes cluster on the cloud selected by `host`. Set exactly one of `uuid`, `billing_uuid`, or `name`. A `PENDING` cluster may have an empty kubeconfig.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Service UUID of the cluster. Exactly one of uuid, billing_uuid, or name must be set.",
				Validators:  ids,
			},
			"billing_uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Portal billing UUID of the cluster.",
				Validators:  ids,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cluster name as returned by the API, or the Terraform-facing name. Not unique.",
				Validators:  ids,
			},
			"region":             schema.StringAttribute{Optional: true, Computed: true, Description: "Optional region filter (API field location), also populated from the API."},
			"creator_uuid":       schema.StringAttribute{Computed: true, Description: "UUID of the account that created the cluster."},
			"cluster_name":       schema.StringAttribute{Computed: true, Description: "Kubernetes cluster name."},
			"cluster_status":     schema.StringAttribute{Computed: true, Description: "Cluster status from Meson (for example PENDING or RUNNING)."},
			"endpoint":           schema.StringAttribute{Computed: true, Description: "Kubernetes API endpoint."},
			"kube_config":        schema.StringAttribute{Computed: true, Sensitive: true, Description: "Kubeconfig. Empty while the cluster is still pending."},
			"kubernetes_version": schema.StringAttribute{Computed: true, Description: "Kubernetes version."},
			"network_cidr":       schema.StringAttribute{Computed: true, Description: "Cluster network CIDR."},
			"status":             schema.StringAttribute{Computed: true, Description: "Portal billing status, when billing_uuid is available."},
		},
	}
}

func (d *kaasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSourceClient(req, resp)
}

func (d *kaasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config kaasDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := optionalString(config.UUID)
	billingUUID := optionalString(config.BillingUUID)
	name := optionalString(config.Name)
	region := optionalString(config.Region)

	var found *client.KAASResponse
	if uuid != "" {
		res, err := d.client.GetKAASResource(uuid)
		if err != nil {
			resp.Diagnostics.AddError("Error Reading KAAS Resource", "Could not read kaas resource: "+err.Error())
			return
		}
		if region != "" && res.Region != region {
			resp.Diagnostics.AddError("No matching elemento_kaas", "The cluster exists but is not in the requested region.")
			return
		}
		found = res
	} else {
		resources, err := d.client.ListKAASResources()
		if err != nil {
			resp.Diagnostics.AddError("Error Reading KAAS Resources", "Could not list kaas resources: "+err.Error())
			return
		}
		resources = filterByRegion(resources, region, func(item client.KAASResponse) string { return item.Region })

		var matches []client.KAASResponse
		for _, item := range resources {
			if billingUUID != "" && item.BillingUUID == billingUUID {
				matches = append(matches, item)
				continue
			}
			if name != "" && (item.Name == name || item.ClusterName == name) {
				matches = append(matches, item)
			}
		}
		one, diags := requireOneMatch(matches, "elemento_kaas")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		found = one
	}

	config.UUID = types.StringValue(found.UUID)
	config.BillingUUID = types.StringValue(found.BillingUUID)
	if found.Name != "" {
		config.Name = types.StringValue(found.Name)
	} else {
		config.Name = types.StringValue(found.ClusterName)
	}
	config.Region = types.StringValue(found.Region)
	config.CreatorUUID = types.StringValue(found.CreatorUUID)
	config.ClusterName = types.StringValue(found.ClusterName)
	config.ClusterStatus = types.StringValue(found.Status)
	config.Endpoint = types.StringValue(found.Endpoint)
	config.KubeConfig = types.StringValue(found.KubeConfig)
	config.Version = types.StringValue(found.Version)
	config.NetworkCIDR = types.StringValue(found.NetworkCIDR)

	status, diags := resourceStatus(d.client, found.BillingUUID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Status = status

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
