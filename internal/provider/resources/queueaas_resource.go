package resources

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &queueAASResource{}
	_ resource.ResourceWithConfigure   = &queueAASResource{}
	_ resource.ResourceWithImportState = &queueAASResource{}
)

func NewQueueAASResource() resource.Resource {
	return &queueAASResource{}
}

type queueAASResource struct {
	common.BaseResource
}

func (r *queueAASResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_queueaas"
}

func (r *queueAASResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["node_number"] = schema.Int64Attribute{
		Required:    true,
		Description: "The number of nodes of the queue.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["node_size"] = schema.StringAttribute{
		Required:    true,
		Description: "The size of each node of the queue.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["size_gb"] = schema.Int64Attribute{
		Required:    true,
		Description: "The size (in GB) of the queue.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplace(),
		},
	}
	attributes["username"] = schema.StringAttribute{
		Required:    true,
		Description: "The username to use for the queue.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["password"] = schema.StringAttribute{
		Required:    true,
		Description: "The password to use for the queue.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Sensitive: true,
	}
	attributes["kafka_version"] = schema.StringAttribute{
		Required:    true,
		Description: "The version of Kafka to use for the queue.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["is_ha"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whether HA should be enabled for the resource or not.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["is_public"] = schema.BoolAttribute{
		Required:    true,
		Description: "Whetherthe resource should be public or not.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
	attributes["network_uuid"] = schema.StringAttribute{
		Required:    true,
		Description: "The ID of the model used by the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attributes["bootstrap_servers"] = schema.StringAttribute{
		Computed:    true,
		Description: "The boostrap servers of the resource.",
	}
	attributes["kafka_cert"] = schema.StringAttribute{
		Computed:    true,
		Description: "The Kafka certificate for the resource.",
		Sensitive:   true,
	}

	resp.Schema = schema.Schema{
		Attributes: attributes,
	}
}

func (r *queueAASResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan common.QueueAASResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queueaasResource := client.QueueAASRequest{
		BaseRequest: client.BaseRequest{
			Region:           plan.Region.ValueString(),
			BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
		},
		Name:         plan.Name.ValueString(),
		NodeNumber:   plan.NodeNumber.ValueInt64(),
		NodeSize:     plan.NodeSize.ValueString(),
		SizeGB:       plan.SizeGB.ValueInt64(),
		Username:     plan.Username.ValueString(),
		Password:     plan.Password.ValueString(),
		KafkaVersion: plan.KafkaVersion.ValueString(),
		IsHA:         plan.IsHA.ValueBool(),
		IsPublic:     plan.IsPublic.ValueBool(),
		NetworkUUID:  plan.NetworkUUID.ValueString(),
	}

	canCreateResponse, err := r.Client.CanCreateResource(ctx, queueaasResource)
	if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "QueueAAS", err) {
		return
	}

	response, err := r.Client.CreateResource(ctx, queueaasResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating QueueAAS Resource",
			"Could not create queueaas resource, unexpected error: "+err.Error(),
		)
		return
	}

	switch res := response.(type) {
	case client.PublicCreateResponse:
		plan.BillingUUID = types.StringValue(res.BillingUUID)
		plan.PaymentURL = types.StringValue(res.PaymentURL)
		plan.UUID = types.StringNull()

		diags = resp.State.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)

		status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Reading QueueAAS Resource Status",
				"Could not read queueaas resource status: "+err.Error(),
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
			fmt.Sprintf("CreateQueueAASResource returned unexpected type %T", response),
		)
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	plan.CreationDate = types.StringNull()
	plan.CreatorUUID = types.StringNull()
	plan.BootstrapServers = types.StringNull()
	plan.KafkaCert = types.StringNull()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *queueAASResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model common.QueueAASResourceModel
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

func (r *queueAASResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *queueAASResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state common.QueueAASResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UUID.IsNull() {
		resp.Diagnostics.AddError(
			"Error Deleting QueueAAS Resource",
			"Could not delete queueaas resource: resource uuid is null.",
		)
		return
	}

	statusCode, err := r.Client.DeleteResource(ctx, state.UUID.ValueString(), "queueaas")
	if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting QueueAAS Resource",
				"Could not delete queueaas resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting QueueAAS Resource",
				"Could not delete queueaas resource, unexpected error: "+err.Error(),
			)
		}
		return
	}
}

func (r *queueAASResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
