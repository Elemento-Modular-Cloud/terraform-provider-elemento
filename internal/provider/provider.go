package provider

import (
	"context"
	"os"
	"terraform-provider-elemento/internal/client"

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
		Description: "The Elemento provider manages infrastructure across every Atomosphere-connected cloud through a single Terraform schema.",
		MarkdownDescription: "The Elemento provider is the Terraform interface to [Elemento Modular Cloud](https://elemento.cloud). " +
			"It replaces the *many clouds, many providers* model with **one provider** and a destination cloud passed as an argument.\n\n" +
			"Set `target` to the [Atomosphere](https://elemento.cloud/atomosphere.html) target name for the backend you intend to use. " +
			"Resource types stay the same across clouds. Use provider aliases to target more than one cloud in a single root module.\n\n" +
			"Arguments may also be supplied with `ELEMENTO_TARGET`, `ELEMENTO_USERNAME`, `ELEMENTO_PASSWORD`, and `ELEMENTO_PORTAL`.\n\n" +
			"```terraform\n" +
			"provider \"elemento\" {\n" +
			"  target     = \"azure.dev.meson.elemento.cloud\"\n" +
			"  username = var.elemento_username\n" +
			"  password = var.elemento_password\n" +
			"}\n" +
			"```\n",
		Attributes: map[string]schema.Attribute{
			"target": schema.StringAttribute{
				Optional:    true,
				Description: "The elemento API target name. Can also be set via the ELEMENTO_TARGET environment variable.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "The Elemento API username. Can also be set via the ELEMENTO_USERNAME environment variable.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Elemento API password. Can also be set via the ELEMENTO_PASSWORD environment variable.",
			},
			"portal": schema.StringAttribute{
				Optional:    true,
				Description: "Portal hostname for authentication and billing. Defaults to portal.elemento.cloud. Can also be set via the ELEMENTO_PORTAL environment variable.",
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
		NewComputeDataSource,
		NewStorageDataSource,
		NewObjectStorageDataSource,
		NewDBAASDataSource,
		NewKAASDataSource,
	}
}

func (p *ElementoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewComputeResource,
		NewStorageResource,
		NewObjectStorageResource,
		NewDBAASResource,
		NewKAASResource,
		NewN8NResource,
		NewOpenClawResource,
		NewFileStorageResource,
		NewPrivateNetworkResource,
	}
}
