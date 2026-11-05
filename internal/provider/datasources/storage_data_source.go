package datasources

import (
	"context"

	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &storageDataSource{}
	_ datasource.DataSourceWithConfigure = &storageDataSource{}
)

func NewStorageDataSource() datasource.DataSource {
	return &storageDataSource{}
}

type storageDataSource struct {
	common.BaseResource
}

type storageDataSourceModel struct {
	VolumeUUID  types.String `tfsdk:"volume_uuid"`
	BillingUUID types.String `tfsdk:"billing_uuid"`
	Name        types.String `tfsdk:"name"`
	Region      types.String `tfsdk:"region"`
	CreatorUUID types.String `tfsdk:"creator_uuid"`
	Size        types.Int64  `tfsdk:"size"`
	Private     types.Bool   `tfsdk:"private"`
	Shareable   types.Bool   `tfsdk:"shareable"`
	Bootable    types.Bool   `tfsdk:"bootable"`
	Readonly    types.Bool   `tfsdk:"readonly"`
	Clonable    types.Bool   `tfsdk:"clonable"`
	Priority    types.Int32  `tfsdk:"priority"`
	Tags        types.List   `tfsdk:"tags"`
	Status      types.String `tfsdk:"status"`
}

func (d *storageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func (d *storageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ids := identifierExactlyOneOf("volume_uuid", "billing_uuid", "name")
	resp.Schema = schema.Schema{
		Description:         "Looks up an existing Elemento block volume on the cloud selected by host.",
		MarkdownDescription: "Looks up an existing Elemento block volume on the cloud selected by `host`. Prefer `volume_uuid`. Set exactly one of `volume_uuid`, `billing_uuid`, or `name`. Size is in MB as documented by Meson.",
		Attributes: map[string]schema.Attribute{
			"volume_uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Volume UUID. Preferred lookup key. Exactly one of volume_uuid, billing_uuid, or name must be set.",
				Validators:  ids,
			},
			"billing_uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Portal billing UUID. Best-effort; Meson list/get payloads may omit it.",
				Validators:  ids,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Volume name. Not unique; use volume_uuid when more than one match exists.",
				Validators:  ids,
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional region filter, also populated from the API.",
			},
			"creator_uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Account UUID that owns the volume. Defaults to the authenticated account.",
			},
			"size":      schema.Int64Attribute{Computed: true, Description: "Volume size in MB, as documented by Meson."},
			"private":   schema.BoolAttribute{Computed: true, Description: "Whether the volume is private."},
			"shareable": schema.BoolAttribute{Computed: true, Description: "Whether the volume is shareable."},
			"bootable":  schema.BoolAttribute{Computed: true, Description: "Whether the volume is bootable."},
			"readonly":  schema.BoolAttribute{Computed: true, Description: "Whether the volume is read-only."},
			"clonable":  schema.BoolAttribute{Computed: true, Description: "Whether the volume is clonable."},
			"priority":  schema.Int32Attribute{Computed: true, Description: "Boot priority."},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the volume.",
			},
			"status": schema.StringAttribute{Computed: true, Description: "Portal billing status, when billing_uuid is available."},
		},
	}
}

func (d *storageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Client = configureDataSourceClient(req, resp)
}

func (d *storageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config storageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumeUUID := optionalString(config.VolumeUUID)
	billingUUID := optionalString(config.BillingUUID)
	name := optionalString(config.Name)
	region := optionalString(config.Region)
	creatorUUID := optionalString(config.CreatorUUID)

	if creatorUUID == "" {
		id, err := d.Client.ClientUUID()
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Account UUID", "Could not resolve creator_uuid: "+err.Error())
			return
		}
		creatorUUID = *id
	}

	var found *client.StorageResourceResponse
	if volumeUUID != "" {
		res, err := d.Client.GetStorageResource(creatorUUID, volumeUUID)
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Storage Resource", "Could not read storage resource: "+err.Error())
			return
		}
		if region != "" && res.Region != region {
			resp.Diagnostics.AddError("No matching elemento_storage", "The volume exists but is not in the requested region.")
			return
		}
		found = res
	} else {
		resources, err := d.Client.ListStorageResources(creatorUUID)
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Storage Resources", "Could not list storage resources: "+err.Error())
			return
		}
		resources = filterByRegion(resources, region, func(item client.StorageResourceResponse) string { return item.Region })

		var matches []client.StorageResourceResponse
		for _, item := range resources {
			if billingUUID != "" && item.BillingUUID == billingUUID {
				matches = append(matches, item)
				continue
			}
			if name != "" && item.Name == name {
				matches = append(matches, item)
			}
		}
		one, diags := requireOneMatch(matches, "elemento_storage")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		found = one
	}

	config.VolumeUUID = types.StringValue(found.UUID)
	if found.BillingUUID != "" {
		config.BillingUUID = types.StringValue(found.BillingUUID)
	} else {
		config.BillingUUID = types.StringNull()
	}
	config.Name = types.StringValue(found.Name)
	config.Region = types.StringValue(found.Region)
	config.CreatorUUID = types.StringValue(found.CreatorUUID)
	if found.CreatorUUID == "" {
		config.CreatorUUID = types.StringValue(creatorUUID)
	}
	config.Size = types.Int64Value(int64(found.Size))
	config.Private = types.BoolValue(found.Private)
	config.Shareable = types.BoolValue(found.Shareable)
	config.Bootable = types.BoolValue(found.Bootable)
	config.Readonly = types.BoolValue(found.Readonly)
	config.Clonable = types.BoolValue(found.Clonable)
	config.Priority = types.Int32Value(int32(found.Priority))

	tags, diags := types.ListValueFrom(ctx, types.StringType, found.Tags)
	resp.Diagnostics.Append(diags...)
	if found.Tags == nil {
		tags = types.ListValueMust(types.StringType, []attr.Value{})
	}
	config.Tags = tags

	status, diags := resourceStatus(d.Client, found.BillingUUID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Status = status

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
