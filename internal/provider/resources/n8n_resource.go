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
	_ resource.Resource                = &n8nResource{}
	_ resource.ResourceWithConfigure   = &n8nResource{}
	_ resource.ResourceWithImportState = &n8nResource{}
)

func NewN8NResource() resource.Resource {
	return &n8nResource{}
}

type n8nResource struct {
	common.BaseResource
}

func (r *n8nResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_n8n"
}

func (r *n8nResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema(ctx)

	attributes["vm_name"] = schema.StringAttribute{
		Required:    true,
		Description: "Name of the virtual machine running n8n.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["cpu_cores"] = schema.Int32Attribute{
		Required:    true,
		Description: "Number of CPU cores.",
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.RequiresReplace(),
		},
	}
	attributes["ram_size_mb"] = schema.Int32Attribute{
		Required:    true,
		Description: "RAM in MB.",
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.RequiresReplace(),
		},
	}
	attributes["disk_size_gb"] = schema.Int64Attribute{
		Required:    true,
		Description: "Boot disk size in GB.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["username"] = schema.StringAttribute{
		Required:    true,
		Description: "Username of the VM user.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["password"] = schema.StringAttribute{
		Required:    true,
		Sensitive:   true,
		Description: "Password of the VM user.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["ssh_key"] = schema.StringAttribute{
		Optional:    true,
		Sensitive:   true,
		Description: "SSH public key authorized on the VM.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["service_url"] = schema.StringAttribute{
		Computed:    true,
		Description: "URL of the n8n instance.",
	}
	attributes["state"] = schema.StringAttribute{
		Computed:    true,
		Description: "Deployment state, e.g. `ERROR` or `SUCCESS`.",
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *n8nResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *n8nResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.N8NResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cloudinit, err := r.Client.RenderCloudinit(ctx, "n8n", map[string]string{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating n8n Resource",
			"Could not render cloudinit template: "+err.Error(),
		)
		return
	}
	cloudinit_b64 := base64.StdEncoding.EncodeToString([]byte(*cloudinit))

	n8nResource := client.SpotVMRequest{
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
				Name:   "n8n-main-disk",
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
		Tags:            []string{"service:n8n"},
		CloudInitBase64: cloudinit_b64,
	}

	canCreateResponse, err := r.Client.CanCreateResource(ctx, n8nResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "n8n", err) {
		return
	}

	response, err := r.Client.CreateResource(ctx, n8nResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating n8n Resource",
			"Could not create n8n resource, unexpected error: "+err.Error(),
		)
		return
	}

	mustRead := false
	switch res := response.(type) {
	case client.PublicCreateResponse:
		plan.BillingUUID = types.StringValue(res.BillingUUID)
		plan.PaymentURL = types.StringValue(res.PaymentURL)
		plan.UUID = types.StringNull()

		status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Reading n8n Resource Status",
				"Could not read n8n resource status: "+err.Error(),
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

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if mustRead {
    	stateConf := common.GetCreateStateConf(ctx, r.Client, createTimeout, &plan)

		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			resp.Diagnostics.AddError("Timeout waiting for resource creation", err.Error())
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *n8nResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.N8NResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, _, diags := model.Read(ctx, r.Client, common.RemoveResource(resp))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || !found {
		return
	}
	model.PaymentURL = types.StringNull()

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

func (r *n8nResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *n8nResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.N8NResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error Deleting n8n Resource",
			"Could not delete n8n resource: resource uuid is null.",
		)
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusCode, err := r.Client.DeleteSpotVMResource(ctx, state.UUID.ValueString())
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting n8n Resource",
				"Could not delete n8n resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting n8n Resource",
				"Could not delete n8n resource, unexpected error: "+err.Error(),
			)
		}
		return
	}

	stateConf := common.GetDeleteStateConf(ctx, r.Client, deleteTimeout, &state)
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Resource Deletion",
			fmt.Sprintf("Could not confirm deletion of resource %s: %s", state.UUID.ValueString(), err.Error()),
		)
		return
	}
}

func (r *n8nResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
