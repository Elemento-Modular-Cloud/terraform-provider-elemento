package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &fileStorageResource{}
	_ resource.ResourceWithConfigure = &fileStorageResource{}
)

func NewFileStorageResource() resource.Resource {
	return &fileStorageResource{}
}

type fileStorageResource struct {
	client *client.ElementoClient
}

type fileStorageResourceModel struct {
	baseResourceModel
	Name         types.String  `tfsdk:"name"`
	SizeGB       types.Int64   `tfsdk:"size_gb"`
	UsedGB       types.Float64 `tfsdk:"used_gb"`
	IsPublic     types.Bool    `tfsdk:"is_public"`
	DNSName      types.String  `tfsdk:"dns_name"`
	State        types.String  `tfsdk:"state"`
	NFSHost      types.String  `tfsdk:"nfs_host"`
	NFSPath      types.String  `tfsdk:"nfs_path"`
	NFSVersion   types.String  `tfsdk:"nfs_version"`
	MountCommand types.String  `tfsdk:"mount_command"`
}

func (r *fileStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filestorage"
}

func (r *fileStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := baseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Required:    true,
		Description: "The size (in GB) of the resource.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["is_public"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether this resource is public or not.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["used_gb"] = schema.Float64Attribute{
		Computed:    true,
		Description: "The size (in GB) currently used by the resource.",
	}
	attributes["dns_name"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the DNS associated with the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the resource.",
	}
	attributes["nfs_host"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs host associated with the resource",
	}
	attributes["nfs_path"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs path associated with the resource",
	}
	attributes["nfs_version"] = schema.StringAttribute{
		Computed:    true,
		Description: "The nfs version associated with the resource",
	}
	attributes["mount_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "The command used to connect to the resource",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *fileStorageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ElementoClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *ElementoClient, got %T. Please report this issue to the provider developer.", req.ProviderData),
		)

		return
	}
	r.client = client
}

func (r *fileStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fileStorageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileStorageResource := client.FileStorageRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:   plan.Name.ValueString(),
		SizeGB: int64(plan.SizeGB.ValueInt64()),
	}

	canCreateResponse, err := r.client.CanCreateFileStorageResource(fileStorageResource)
    ValidateCanCreate(canCreateResponse, &diags, "FileStorage", err)

	response, err := r.client.CreateFileStorageResource(fileStorageResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating FileStorage Resource",
			"Could not create filestorage resource, unexpected error: "+err.Error(),
		)
		return
	}

	switch res := response.(type) {
	case client.PublicCreateResponse:
		plan.BillingUUID = types.StringValue(res.BillingUUID)
		plan.PaymentURL = types.StringValue(res.PaymentURL)
		plan.UUID = types.StringNull()

		status, err := r.client.GetResourceStatus(res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading FileStorage Resource Status",
				"Could not read filestorage resource status: "+err.Error(),
			)
			return
		}

		plan.Status = types.StringValue(*status)
	case client.PrivateCreateResponse:
		plan.BillingUUID = types.StringNull()
		plan.PaymentURL = types.StringNull()
		plan.UUID = types.StringValue(res.ServiceUUID)
		plan.Status = types.StringValue(res.Status)
	default:
		resp.Diagnostics.AddError(
			"Unexpected API Response Type",
			fmt.Sprintf("CreateFileStorageResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.UsedGB = types.Float64Null()
	plan.DNSName = types.StringNull()
	plan.State = types.StringNull()
	plan.NFSHost = types.StringNull()
	plan.NFSPath = types.StringNull()
	plan.NFSVersion = types.StringNull()
	plan.MountCommand = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *fileStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fileStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resource *client.FileStorageResponse
	if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
		resources, err := r.client.ListFileStorageResources()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading FileStorage Resources",
				"Could not list filestorage resources: "+err.Error(),
			)
			return
		}

		for _, res := range resources {
			if res.BillingUUID == state.BillingUUID.ValueString() {
				resource = &res
				break
			}
		}
	} else {
		res, code, err := r.client.GetFileStorageResource(state.UUID.ValueString())
		if err != nil && code != 404 {
			resp.Diagnostics.AddError(
				"Error Reading FileStorage Resource"+string(code),
				"Could not read filestorage resource: "+err.Error(),
			)
			return
		}
		resource = res
	}

	billing_uuid := state.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		var err error
		status, err = r.client.GetResourceStatus(state.BillingUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Compute Resource Status",
				"Could not read compute resource status: "+err.Error(),
			)
			return
		}
	} else {
		status = nil
	}

	if resource == nil && (status == nil || *status == "deleted" || *status == "error" || *status == "canceled") {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Status = types.StringPointerValue(status)
	state.Name = types.StringValue(resource.Name)
	state.Region = types.StringValue(resource.Region)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	state.UUID = types.StringValue(resource.ServiceUUID)
	state.CreatorUUID = types.StringValue(resource.CreatorUUID)
	state.CreationDate = types.StringValue(resource.CreationDate)
	state.SizeGB = types.Int64Value(resource.SizeGB)
	state.UsedGB = types.Float64Value(resource.UsedGB)
	state.DNSName = types.StringValue(resource.DNSName)
	state.State = types.StringValue(resource.State)
	state.PaymentURL = types.StringNull()

	credentials, err := r.client.GetFileStorageCredentials(resource.ServiceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading FileStorage Credentials",
			"Could not read filestorage credentials: "+err.Error(),
		)
		return
	}

	state.NFSHost = types.StringValue(credentials.NFSHost)
	state.NFSPath = types.StringValue(credentials.NFSPath)
	state.NFSVersion = types.StringValue(credentials.NFSVersion)
	state.MountCommand = types.StringValue(credentials.MountCommand)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *fileStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *fileStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fileStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFileStorageResource(state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting FileStorage Resource",
			"Could not delete filestorage resource, unexpected error: "+err.Error(),
		)
		return
	}
}
