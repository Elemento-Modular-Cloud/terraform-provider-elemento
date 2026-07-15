package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &storageResource{}
	_ resource.ResourceWithConfigure = &storageResource{}
)

func NewStorageResource() resource.Resource {
	return &storageResource{}
}

type storageResource struct{
	client *ElementoClient
}

type storageVolumeModel struct {
    Name        types.String `tfsdk:"name"`
    Size        types.Int64  `tfsdk:"size"`
    Private     types.Bool   `tfsdk:"private"`
    Shareable   types.Bool   `tfsdk:"shareable"`
    Bootable    types.Bool   `tfsdk:"bootable"`
    Readonly    types.Bool   `tfsdk:"readonly"`
    Priority    types.Int32  `tfsdk:"priority"`
    VolumeUUID  types.String `tfsdk:"volume_uuid"`
    ServerURL   types.String `tfsdk:"serverurl"`
    LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *storageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute"
}

func volumeAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "name": schema.StringAttribute{
            Required: true,
        },
        "size": schema.Int64Attribute{
            Required: true,
        },
        "private": schema.BoolAttribute{
	       	Required: true,
        },
        "shareable": schema.BoolAttribute{
	       	Required: true,
        },
        "bootable": schema.BoolAttribute{
	       	Required: true,
        },
        "readonly": schema.BoolAttribute{
	       	Required: true,
        },
        "priority": schema.Int32Attribute{
	       	Computed: true,
        },
        "volume_uuid": schema.StringAttribute{
 	       	Computed: true,
        },
        "serverurl": schema.StringAttribute{
	       	Computed: true,
        },
        "last_updates": schema.StringAttribute{
        	Computed: true,
        },
   }
}

func (r *storageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: volumeAttributes(),
	}
}

func (r *storageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ElementoClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *ElementoClient, got %T. Please report this issue to the provider developer.", req.ProviderData),
		)

		return
	}
	r.client = client
}

func (r *storageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *storageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *storageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *storageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
