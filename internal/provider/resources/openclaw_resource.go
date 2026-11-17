package resources

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &openclawResource{}
	_ resource.ResourceWithConfigure   = &openclawResource{}
	_ resource.ResourceWithImportState = &openclawResource{}
)

func NewOpenClawResource() resource.Resource {
	return &openclawResource{}
}

type openclawResource struct {
	common.BaseResource
}

func (r *openclawResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openclaw"
}

func (r *openclawResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["vm_name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the virtual machine for the openclaw resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["cpu_cores"] = schema.Int32Attribute{
		Required:    true,
		Description: "The number of CPU cores for the openclaw resource.",
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.RequiresReplace(),
		},
	}
	attributes["ram_size_mb"] = schema.Int32Attribute{
		Required:    true,
		Description: "The amount of RAM in MB for the openclaw resource.",
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.RequiresReplace(),
		},
	}
	attributes["disk_size_gb"] = schema.Int64Attribute{
		Required:    true,
		Description: "The size (in GB) of the boot disk for the openclaw resource.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["username"] = schema.StringAttribute{
		Required:    true,
		Description: "The username for the openclaw resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["password"] = schema.StringAttribute{
		Required:    true,
		Sensitive:   true,
		Description: "The password for the openclaw resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["ssh_key"] = schema.StringAttribute{
		Optional:    true,
		Sensitive:   true,
		Description: "The SSH key for the openclaw resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["service_url"] = schema.StringAttribute{
		Computed:    true,
		Description: "The service URL for the openclaw resource, assigned after creation.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "The state of the openclaw resource (e.g., 'ERROR', 'SUCCESS').",
	}

	attributes["token"] = schema.StringAttribute{
		Computed:    true,
		Sensitive:   true,
		Description: "The token for the openclaw resource, assigned after creation.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *openclawResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.Client = client
}

func (r *openclawResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.OpenClawResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cloudinit, err := r.Client.RenderCloudinit(ctx, "openclaw", map[string]string{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating openclaw Resource",
			"Could not render cloudinit template: "+err.Error(),
		)
		return
	}
	cloudinit_b64 := base64.StdEncoding.EncodeToString([]byte(*cloudinit))

	openclawResource := client.SpotVMRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		VMName: plan.VMName.ValueString(),
		CPU: client.SpotVMCPU{
			Arch:        []string{"X86_64"},
			Slots:       int(plan.CPUCores.ValueInt32()),
			SharedCores: false,
		},
		Mem: client.SpotVMMemory{
			CapacityMB: int(plan.RAMSizeMB.ValueInt32()),
		},
		OS: client.SpotVMOS{
			Family:  "linux",
			Flavour: "ubuntu",
			Version: "24.04",
		},
		Storage: client.SpotVMStorage{
			Boot: client.SpotVMVolume{
				Name:   "openclaw-main-disk",
				SizeGB: float64(plan.DiskSizeGB.ValueInt64()),
			},
			Data: []client.SpotVMVolume{},
		},
		PCI: client.SpotVMPCI{
			Devices: []client.SpotVMPCIDev{},
		},
		Networks: []client.SpotVMNetwork{
			{Kind: "natted"},
		},
		Auth: client.SpotVMAuth{
			Username: plan.Username.ValueString(),
			Password: plan.Password.ValueString(),
			SSHKey:   plan.SSHKey.ValueStringPointer(),
		},
		Tags:            []string{"service:openclaw"},
		CloudInitBase64: cloudinit_b64,
	}

	canCreateResponse, err := r.Client.CanCreateResource(ctx, openclawResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "OpenClaw", err) {
		return
	}

	response, err := r.Client.CreateResource(ctx, openclawResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating OpenClaw Resource",
			"Could not create openclaw resource, unexpected error: "+err.Error(),
		)
		return
	}

	switch res := response.(type) {
	case client.PublicCreateResponse:
		plan.BillingUUID = types.StringValue(res.BillingUUID)
		plan.PaymentURL = types.StringValue(res.PaymentURL)
		plan.UUID = types.StringNull()

		status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Reading OpenClaw Resource Status",
				"Could not read openclaw resource status: "+err.Error(),
			)
			status = nil
		}

		plan.Status = types.StringPointerValue(status)
	case client.PrivateCreateResponse:
		plan.BillingUUID = types.StringNull()
		plan.PaymentURL = types.StringNull()
		plan.UUID = types.StringValue(res.ServiceUUID)
		plan.Status = types.StringValue(res.Status)
	}

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.State = types.StringNull()
	plan.ServiceURL = types.StringNull()
	plan.Token = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *openclawResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.OpenClawResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := model.Read(ctx, r.Client, common.RemoveResource(resp))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || !found {
		return
	}
	model.PaymentURL = types.StringNull()

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)

}

func (r *openclawResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *openclawResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.OpenClawResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error Deleting OpenClaw Resource",
			"Could not delete openclaw resource: resource uuid is null.",
		)
		return
	}

	statusCode, err := r.Client.DeleteSpotVMResource(ctx, state.UUID.ValueString())
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting OpenClaw Resource",
				"Could not delete openclaw resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting OpenClaw Resource",
				"Could not delete openclaw resource, unexpected error: "+err.Error(),
			)
		}
		return
	}
}

func (r *openclawResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
