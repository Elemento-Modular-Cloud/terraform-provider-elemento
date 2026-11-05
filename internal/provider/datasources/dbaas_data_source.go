package datasources

import (
	"context"

	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &dbaasDataSource{}
	_ datasource.DataSourceWithConfigure = &dbaasDataSource{}
)

func NewDBAASDataSource() datasource.DataSource {
	return &dbaasDataSource{}
}

type dbaasDataSource struct {
	common.BaseResource
}

type dbaasDataSourceModel struct {
	UUID             types.String `tfsdk:"uuid"`
	BillingUUID      types.String `tfsdk:"billing_uuid"`
	Name             types.String `tfsdk:"name"`
	Region           types.String `tfsdk:"region"`
	CreatorUUID      types.String `tfsdk:"creator_uuid"`
	Engine           types.String `tfsdk:"engine"`
	BackupTime       types.String `tfsdk:"backup_time"`
	DiskSizeGB       types.Int64  `tfsdk:"disk_size_gb"`
	NodesNumber      types.Int32  `tfsdk:"nodes_number"`
	CreationDate     types.String `tfsdk:"creation_date"`
	DatabaseHost     types.String `tfsdk:"database_host"`
	DatabasePort     types.String `tfsdk:"database_port"`
	DatabaseUser     types.String `tfsdk:"database_user"`
	DatabasePassword types.String `tfsdk:"database_password"`
	DatabaseName     types.String `tfsdk:"database_name"`
	DatabaseSSLMode  types.String `tfsdk:"database_ssl_mode"`
	PublicURI        types.String `tfsdk:"public_uri"`
	Status           types.String `tfsdk:"status"`
}

func (d *dbaasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaas"
}

func (d *dbaasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ids := identifierExactlyOneOf("uuid", "billing_uuid", "name")
	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento managed database on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento managed database on the cloud selected by `host`. Set exactly one of `uuid`, `billing_uuid`, or `name`. Names are not unique; multiple matches return an error.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Service UUID of the database. Exactly one of uuid, billing_uuid, or name must be set.",
				Validators:  ids,
			},
			"billing_uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Portal billing UUID of the database.",
				Validators:  ids,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Database name. Not unique; use uuid when more than one match exists.",
				Validators:  ids,
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional region filter, also populated from the API.",
			},
			"creator_uuid":      schema.StringAttribute{Computed: true, Description: "UUID of the account that created the database."},
			"engine":            schema.StringAttribute{Computed: true, Description: "Database engine (for example postgres or mysql)."},
			"backup_time":       schema.StringAttribute{Computed: true, Description: "Daily backup time."},
			"disk_size_gb":      schema.Int64Attribute{Computed: true, Description: "Disk size as returned by the API."},
			"nodes_number":      schema.Int32Attribute{Computed: true, Description: "Number of database nodes."},
			"creation_date":     schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"database_host":     schema.StringAttribute{Computed: true, Description: "Database host."},
			"database_port":     schema.StringAttribute{Computed: true, Description: "Database port."},
			"database_user":     schema.StringAttribute{Computed: true, Description: "Database username."},
			"database_password": schema.StringAttribute{Computed: true, Sensitive: true, Description: "Database password."},
			"database_name":     schema.StringAttribute{Computed: true, Description: "Database name used for connections."},
			"database_ssl_mode": schema.StringAttribute{Computed: true, Description: "SSL mode for the database connection."},
			"public_uri":        schema.StringAttribute{Computed: true, Description: "Public connection URI."},
			"status":            schema.StringAttribute{Computed: true, Description: "Portal billing status, when billing_uuid is available."},
		},
	}
}

func (d *dbaasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Client = configureDataSourceClient(req, resp)
}

func (d *dbaasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dbaasDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := optionalString(config.UUID)
	billingUUID := optionalString(config.BillingUUID)
	name := optionalString(config.Name)
	region := optionalString(config.Region)

	var found *client.DBAASResponse
	if uuid != "" {
		res, _, err := d.Client.GetDBAASResource(uuid)
		if err != nil {
			resp.Diagnostics.AddError("Error Reading DBAAS Resource", "Could not read dbaas resource: "+err.Error())
			return
		}
		if region != "" && res.Region != region {
			resp.Diagnostics.AddError("No matching elemento_dbaas", "The database exists but is not in the requested region.")
			return
		}
		found = res
	} else {
		resources, err := d.Client.ListDBAASResources()
		if err != nil {
			resp.Diagnostics.AddError("Error Reading DBAAS Resources", "Could not list dbaas resources: "+err.Error())
			return
		}
		resources = filterByRegion(resources, region, func(item client.DBAASResponse) string { return item.Region })

		var matches []client.DBAASResponse
		for _, item := range resources {
			if billingUUID != "" && item.BillingUUID == billingUUID {
				matches = append(matches, item)
				continue
			}
			if name != "" && item.Name == name {
				matches = append(matches, item)
			}
		}
		one, diags := requireOneMatch(matches, "elemento_dbaas")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		found = one
	}

	config.UUID = types.StringValue(found.UUID)
	config.BillingUUID = types.StringValue(found.BillingUUID)
	config.Name = types.StringValue(found.Name)
	config.Region = types.StringValue(found.Region)
	config.CreatorUUID = types.StringValue(found.CreatorUUID)
	config.Engine = types.StringValue(found.Engine)
	config.BackupTime = types.StringValue(found.BackupTime)
	config.DiskSizeGB = types.Int64Value(found.DiskSizeGB)
	config.NodesNumber = types.Int32Value(found.NodesNumber)
	config.CreationDate = types.StringValue(found.CreationDate)

	status, diags := resourceStatus(d.Client, found.BillingUUID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Status = status

	credentials, err := d.Client.GetDBAASCredentials(found.UUID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading DBAAS Credentials", "Could not read dbaas credentials: "+err.Error())
		return
	}
	config.DatabaseHost = types.StringValue(credentials.DatabaseHost)
	config.DatabasePort = types.StringValue(credentials.DatabasePort)
	config.DatabaseUser = types.StringValue(credentials.DatabaseUser)
	config.DatabasePassword = types.StringValue(credentials.DatabasePassword)
	config.DatabaseName = types.StringValue(credentials.DatabaseName)
	config.DatabaseSSLMode = types.StringValue(credentials.DatabaseSSLMode)
	config.PublicURI = types.StringValue(credentials.PublicURI)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
