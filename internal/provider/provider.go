package provider

import (
	"context"
	"os"
	"terraform-provider-elemento/internal/client"

	"terraform-provider-elemento/internal/provider/resources"
	"terraform-provider-elemento/internal/provider/datasources"


	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &ElementoProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ElementoProvider{
			version: version,
		}
	}
}

type ElementoProvider struct {
	version string
}

func (p *ElementoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "elemento"
	resp.Version = p.version
}

func (p *ElementoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Elemento Modular Cloud resources on any cloud connected through Atomosphere.",
		MarkdownDescription: "Manages [Elemento Modular Cloud](https://elemento.cloud) resources on any cloud connected through [Atomosphere](https://elemento.cloud/atomosphere.html). " +
			"The same resource types work on every cloud; `target` selects which one. To use several clouds in one configuration, declare one provider block per cloud with an `alias`.\n\n" +
			"```terraform\n" +
			"provider \"elemento\" {\n" +
			"  target   = \"azure.dev.meson.elemento.cloud\"\n" +
			"  username = var.elemento_username\n" +
			"  password = var.elemento_password\n" +
			"}\n" +
			"```\n",
		Attributes: map[string]schema.Attribute{
			"target": schema.StringAttribute{
				Optional:    true,
				Description: "Hostname of the Atomosphere (Meson) endpoint of the cloud to manage. Falls back to `ELEMENTO_TARGET`.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Elemento account username. Falls back to `ELEMENTO_USERNAME`.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Elemento account password. Falls back to `ELEMENTO_PASSWORD`.",
			},
			"portal": schema.StringAttribute{
				Optional:    true,
				Description: "Hostname of the Elemento Portal, used for authentication and billing. Defaults to `portal.elemento.cloud`. Falls back to `ELEMENTO_PORTAL`.",
			},
		},
	}
}

type ElementoProviderModel struct {
	Target   types.String `tfsdk:"target"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Portal   types.String `tfsdk:"portal"`
}

func (p *ElementoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ElementoProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the attributes, it must be a known value.

	if config.Target.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("target"),
			"Unknown Elemento API Target",
			"The provider cannot create the Elemento API client as there is an unknown configuration value for the Elemento API target. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ELEMENTO_TARGET environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Elemento API Username",
			"The provider cannot create the Elemento API client as there is an unknown configuration value for the Elemento API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ELEMENTO_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Elemento API Password",
			"The provider cannot create the Elemento API client as there is an unknown configuration value for the Elemento API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ELEMENTO_PASSWORD environment variable.",
		)
	}

	if config.Portal.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("portal"),
			"Unknown Elemento API Portal",
			"The provider cannot create the Elemento API client as there is an unknown configuration value for the Elemento API portal. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ELEMENTO_PORTAL environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	target   := os.Getenv("ELEMENTO_TARGET")
	username := os.Getenv("ELEMENTO_USERNAME")
	password := os.Getenv("ELEMENTO_PASSWORD")
	portal   := os.Getenv("ELEMENTO_PORTAL")

	if !config.Target.IsNull() {
		target = config.Target.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.Portal.IsNull() {
		portal = config.Portal.ValueString()
	}

	if target == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("target"),
			"Missing Elemento API Target",
			"The provider cannot create the Elemento API client as there is a missing or empty value for the Elemento API target. "+
				"Set the target value in the configuration or use the ELEMENTO_TARGET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Elemento API username",
			"The provider cannot create the Elemento API client as there is a missing or empty value for the Elemento API username. "+
				"Set the username value in the configuration or use the ELEMENTO_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Elemento API password",
			"The provider cannot create the Elemento API client as there is a missing or empty value for the Elemento API password. "+
				"Set the password value in the configuration or use the ELEMENTO_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if portal == "" {
		portal = "portal.elemento.cloud"
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := client.NewElementoClient(target, username, password, portal)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ElementoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewSpotVMDataSource,
		datasources.NewStorageDataSource,
		datasources.NewObjectStorageDataSource,
		datasources.NewDBAASDataSource,
		datasources.NewKAASDataSource,
		datasources.NewN8NDataSource,
		datasources.NewOpenClawDataSource,
		datasources.NewFileStorageDataSource,
		datasources.NewPrivateNetworkDataSource,
		datasources.NewPublicIPDataSource,
		datasources.NewRegistryDataSource,
		datasources.NewLoadBalancerDataSource,
		datasources.NewPublicGatewayDataSource,
		datasources.NewServerlessDataSource,
		datasources.NewAIServerlessDataSource,
		datasources.NewQueueAASDataSource,

	}
}

func (p *ElementoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// NewComputeResource,
		resources.NewSpotVMResource,
		resources.NewStorageResource,
		resources.NewObjectStorageResource,
		resources.NewDBAASResource,
		resources.NewKAASResource,
		resources.NewN8NResource,
		resources.NewOpenClawResource,
		resources.NewFileStorageResource,
		resources.NewPrivateNetworkResource,
		resources.NewPublicIPResource,
		resources.NewRegistryResource,
		resources.NewLoadBalancerResource,
		resources.NewPublicGatewayResource,
		resources.NewServerlessResource,
		resources.NewAIServerlessResource,
		resources.NewQueueAASResource,
	}
}
