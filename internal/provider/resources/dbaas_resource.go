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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &dbaasResource{}
	_ resource.ResourceWithConfigure = &dbaasResource{}
	_ resource.ResourceWithImportState = &dbaasResource{}
)

func NewDBAASResource() resource.Resource {
	return &dbaasResource{}
}

type dbaasResource struct{
	common.BaseResource
}

func (r *dbaasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaas"
}

func (r *dbaasResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema(ctx)

    attributes["name"] = schema.StringAttribute{
        Required: true,
        Description: "Name of the database service.",
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.RequiresReplace(),
        },
    }
    attributes["engine"] = schema.StringAttribute{
        Required: true,
        Description: "Database engine, e.g. `postgres` or `mysql`.",
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.RequiresReplace(),
        },
    }
    attributes["backup_time"] = schema.StringAttribute{
        Required: true,
        Description: "Daily backup time, in `HH:MM:SS` format.",
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.RequiresReplace(),
        },
    }
    attributes["disk_size_gb"] = schema.Int64Attribute{
        Required: true,
        Description: "Disk size in GB.",
        PlanModifiers: []planmodifier.Int64{
            int64planmodifier.RequiresReplace(),
        },
    }
    attributes["password"] = schema.StringAttribute{
        Required: true,
        Sensitive: true,
        Description: "Administrator password for the service.",
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.RequiresReplace(),
        },
    }
    attributes["nodes_number"] = schema.Int32Attribute{
        Required: true,
        Description: "Number of database nodes.",
        PlanModifiers: []planmodifier.Int32{
            int32planmodifier.RequiresReplace(),
        },
    }
    attributes["database_host"] = schema.StringAttribute{
        Computed: true,
        Description: "Host to connect to.",
    }
    attributes["database_port"] = schema.StringAttribute{
        Computed: true,
        Description: "Port to connect to.",
    }
    attributes["database_user"] = schema.StringAttribute{
        Computed: true,
        Description: "Username for connections.",
    }
    attributes["database_password"] = schema.StringAttribute{
        Computed: true,
        Sensitive: true,
        Description: "Password for connections.",
    }
    attributes["database_name"] = schema.StringAttribute{
        Computed: true,
        Description: "Name of the database to connect to.",
    }
    attributes["database_ssl_mode"] = schema.StringAttribute{
        Computed: true,
        Description: "SSL mode for connections.",
    }
    attributes["public_uri"] = schema.StringAttribute{
        Computed: true,
        Description: "Public connection URI.",
    }

    resp.Schema = schema.Schema{
        Attributes: attributes,
	}
}

func (r *dbaasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan common.DBAASResourceModel
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

    dbaasResource := client.DBAASRequest{
        Name:             plan.Name.ValueString(),
        Engine:           plan.Engine.ValueString(),
        BackupTime:       plan.BackupTime.ValueStringPointer(),
        DiskSizeGB:       plan.DiskSizeGB.ValueInt64(),
        Password:         plan.Password.ValueString(),
        NodesNumber:      plan.NodesNumber.ValueInt32(),
        Region:           plan.Region.ValueString(),
        BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
    }

    canCreateResponse, err := r.Client.CanCreateDBAASResource(ctx, dbaasResource)
    if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "DBAAS", err) {
	    return
    }

    response, err := r.Client.CreateDBAASResource(ctx, dbaasResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating DBAAS Resource",
            "Could not create dbaas resource, unexpected error: " + err.Error(),
        )
        return
    }

    mustRead := false
    switch res := response.(type) {
        case client.PublicCreateResponse:
            plan.BillingUUID = types.StringValue(res.BillingUUID)
            plan.PaymentURL  = types.StringValue(res.PaymentURL)
            plan.UUID        = types.StringNull()

            status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
            if err != nil {
                resp.Diagnostics.AddWarning(
                    "Error Reading DBAAS Resource Status",
                    "Could not read dbaas resource status: "+err.Error(),
                )
                status = nil
            }

            plan.Status = types.StringPointerValue(status)
        case client.PrivateCreateResponse:
            plan.BillingUUID = types.StringNull()
            plan.PaymentURL  = types.StringNull()
            plan.UUID        = types.StringValue(res.ServiceUUID)
            plan.Status      = types.StringValue(res.Status)

            mustRead = true
		default:
			resp.Diagnostics.AddError(
				"Unexpected API Response Type",
				fmt.Sprintf("CreateDBAASResource returned unexpected type %T", response),
			)
			return
    }

    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    plan.CreationDate = types.StringNull()
    plan.CreatorUUID = types.StringNull()
    plan.DatabaseHost = types.StringNull()
    plan.DatabasePort = types.StringNull()
    plan.DatabaseUser = types.StringNull()
    plan.DatabasePassword = types.StringNull()
    plan.DatabaseName = types.StringNull()
    plan.DatabaseSSLMode = types.StringNull()
    plan.PublicURI = types.StringNull()

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

func (r *dbaasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var model common.DBAASResourceModel
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

func (r *dbaasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *dbaasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state common.DBAASResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    if state.UUID.IsNull() {
	    resp.Diagnostics.AddError(
		    "Error Deleting DBAAS Resource",
		    "Could not delete dbaas resource: resource uuid is null.",
	    )
		return
     }

    statusCode, err := r.Client.DeleteDBAASResource(ctx, state.UUID.ValueString())
    if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting DBAAS Resource",
				"Could not delete dbaas resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting DBAAS Resource",
				"Could not delete dbaas resource, unexpected error: " + err.Error(),
			)
		}
		return
    }
}

func (r *dbaasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
