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
	_ datasource.DataSource              = &objectStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &objectStorageDataSource{}
)

func NewObjectStorageDataSource() datasource.DataSource {
	return &objectStorageDataSource{}
}

type objectStorageDataSource struct {
	common.BaseResource
}

type objectStorageDataSourceModel struct {
	UUID            types.String `tfsdk:"uuid"`
	BillingUUID     types.String `tfsdk:"billing_uuid"`
	Name            types.String `tfsdk:"name"`
	Region          types.String `tfsdk:"region"`
	ActiveStorage   types.Int32  `tfsdk:"active_storage"`
	Endpoint        types.String `tfsdk:"endpoint"`
	AccountName     types.String `tfsdk:"account_name"`
	ContainerName   types.String `tfsdk:"container_name"`
	SasToken        types.String `tfsdk:"sas_token"`
	AzureCliCommand types.String `tfsdk:"azure_cli_command"`
	Status          types.String `tfsdk:"status"`
}

func (d *objectStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_objectstorage"
}

func (d *objectStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ids := identifierExactlyOneOf("uuid", "billing_uuid", "name")
	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento object storage bucket on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento object storage bucket on the cloud selected by `host`. Set exactly one of `uuid`, `billing_uuid`, or `name`.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Service UUID of the bucket. Exactly one of uuid, billing_uuid, or name must be set.",
				Validators:  ids,
			},
			"billing_uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Portal billing UUID of the bucket.",
				Validators:  ids,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Bucket name. Not unique; use uuid when more than one match exists.",
				Validators:  ids,
			},
			"region":            schema.StringAttribute{Optional: true, Computed: true, Description: "Optional region filter, also populated from the API."},
			"active_storage":    schema.Int32Attribute{Computed: true, Description: "Active storage reported by the API."},
			"endpoint":          schema.StringAttribute{Computed: true, Description: "Object storage endpoint."},
			"account_name":      schema.StringAttribute{Computed: true, Description: "Storage account name."},
			"container_name":    schema.StringAttribute{Computed: true, Description: "Container name."},
			"sas_token":         schema.StringAttribute{Computed: true, Sensitive: true, Description: "SAS token for the bucket."},
			"azure_cli_command": schema.StringAttribute{Computed: true, Sensitive: true, Description: "Azure CLI command returned by the credentials endpoint."},
			"status":            schema.StringAttribute{Computed: true, Description: "Portal billing status, when billing_uuid is available."},
		},
	}
}

func (d *objectStorageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Client = configureDataSourceClient(req, resp)
}

func (d *objectStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config objectStorageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := optionalString(config.UUID)
	billingUUID := optionalString(config.BillingUUID)
	name := optionalString(config.Name)
	region := optionalString(config.Region)

	var found *client.ObjectStorageResponse
	if uuid != "" {
		res, err := d.Client.GetObjectStorageResource(uuid)
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Object Storage", "Could not read object storage resource: "+err.Error())
			return
		}
		if region != "" && res.Region != region {
			resp.Diagnostics.AddError("No matching elemento_objectstorage", "The bucket exists but is not in the requested region.")
			return
		}
		found = res
	} else {
		resources, err := d.Client.ListObjectStorageResources()
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Object Storage", "Could not list object storage resources: "+err.Error())
			return
		}
		resources = filterByRegion(resources, region, func(item client.ObjectStorageResponse) string { return item.Region })

		var matches []client.ObjectStorageResponse
		for _, item := range resources {
			if billingUUID != "" && item.BillingUUID == billingUUID {
				matches = append(matches, item)
				continue
			}
			if name != "" && item.Name == name {
				matches = append(matches, item)
			}
		}
		one, diags := requireOneMatch(matches, "elemento_objectstorage")
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
	config.ActiveStorage = types.Int32Value(found.ActiveStorage)
	config.Endpoint = types.StringValue(found.Endpoint)

	status, diags := resourceStatus(d.Client, found.BillingUUID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Status = status

	credentials, err := d.Client.GetObjectStorageCredentials(found.UUID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Object Storage Credentials", "Could not read object storage credentials: "+err.Error())
		return
	}
	config.AccountName = types.StringValue(credentials.AccountName)
	config.ContainerName = types.StringValue(credentials.ContainerName)
	config.SasToken = types.StringValue(credentials.SasToken)
	config.AzureCliCommand = types.StringValue(credentials.AzureCliCommand)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
