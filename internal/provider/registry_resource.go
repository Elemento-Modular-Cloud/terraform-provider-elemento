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
	_ resource.Resource              = &registryResource{}
	_ resource.ResourceWithConfigure = &registryResource{}
)

func NewRegistryResource() resource.Resource {
	return &registryResource{}
}

type registryResource struct {
	baseResource
}

type registryResourceModel struct {
	baseResourceModel
	Name            types.String `tfsdk:"name"`
	SizeGB          types.Int64  `tfsdk:"size_gb"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	State           types.String `tfsdk:"state"`
	RepositoryURI   types.String `tfsdk:"repository_uri"`
	RegistryID      types.String `tfsdk:"registry_id"`
	ImageCount      types.Int64  `tfsdk:"image_count"`
	UsedBytes       types.Int64  `tfsdk:"used_bytes"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	RegistryType    types.String `tfsdk:"registry_type"`
	LoginCommand    types.String `tfsdk:"login_command"`
}

func (r *registryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (r *registryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
		Description: "The size (in GB) of the resource.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["repository_uri"] = schema.StringAttribute{
		Computed:    true,
		Description: "The URI of the repository.",
	}
	attributes["registry_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the registry.",
	}
	attributes["image_count"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The number of images in the registry.",
	}
	attributes["used_bytes"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The size (in bytes) currently used by the resource.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The name of the DNS associated with the resource.",
	}
	attributes["access_key_id"] = schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the access key of the resource.",
	}
	attributes["secret_access_key"] = schema.StringAttribute{
		Computed:    true,
		Description: "The secret access key of the resource.",
	}
	attributes["registry_type"] = schema.StringAttribute{
		Computed:    true,
		Description: "The type of the registry.",
	}
	attributes["login_command"] = schema.StringAttribute{
		Computed:    true,
		Description: "The command used to login to the resource.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *registryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registryResource := client.RegistryRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:     plan.Name.ValueString(),
		SizeGB:   plan.SizeGB.ValueInt64(),
		IsPublic: plan.IsPublic.ValueBool(),
	}

	canCreateResponse, err := r.client.CanCreateResource(registryResource)
	ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "Registry", err)

	response, err := r.client.CreateResource(registryResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Registry Resource",
			"Could not create registry resource, unexpected error: "+err.Error(),
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
				"Error Reading Registry Resource Status",
				"Could not read registry resource status: "+err.Error(),
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
			fmt.Sprintf("CreateRegistryResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.RepositoryURI = types.StringNull()
	plan.RegistryID = types.StringNull()
	plan.ImageCount = types.Int64Null()
	plan.UsedBytes = types.Int64Null()
	plan.AccessKeyID = types.StringNull()
	plan.SecretAccessKey = types.StringNull()
	plan.RegistryType = types.StringNull()
	plan.LoginCommand = types.StringNull()

	plan.State = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *registryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resource *client.RegistryResponse
	if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.RegistryResponse](r.client, "registry")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Registry Resources",
				"Could not list registry resources: "+err.Error(),
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
		res, code, err := client.GetResource[client.RegistryResponse](r.client, state.UUID.ValueString(), "registry")
		if err != nil && code != 404 {
			resp.Diagnostics.AddError(
				"Error Reading Registry Resource"+string(code),
				"Could not read registry resource: "+err.Error(),
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
	state.IsPublic = types.BoolValue(resource.IsPublic)
	state.RepositoryURI = types.StringValue(resource.RepositoryURI)
	state.RegistryID = types.StringValue(resource.RegistryID)
	state.ImageCount = types.Int64Value(resource.ImageCount)
	state.UsedBytes = types.Int64Value(resource.UsedBytes)
	state.State = types.StringValue(resource.State)
	state.PaymentURL = types.StringNull()

	credentials, err := r.client.GetRegistryCredentials(resource.ServiceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Registry Credentials",
			"Could not read registry credentials: "+err.Error(),
		)
		return
	}

	state.AccessKeyID = types.StringValue(credentials.AccessKeyID)
	state.SecretAccessKey = types.StringValue(credentials.SecretAccessKey)
	state.RegistryType = types.StringValue(credentials.RegistryType)
	state.LoginCommand = types.StringValue(credentials.LoginCommand)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *registryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *registryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state registryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResource(state.UUID.ValueString(), "registry")
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Deleting Registry Resource",
			"Could not delete registry resource, unexpected error: "+err.Error(),
		)
		return
	}
}
