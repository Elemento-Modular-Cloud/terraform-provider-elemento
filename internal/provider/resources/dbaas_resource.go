package resources

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"
	"time"

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

func (r *dbaasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "creator_uuid": schema.StringAttribute{
                Computed: true,
                Description: "The UUID of the creator of the resource.",
            },
            "name": schema.StringAttribute{
                Required: true,
                Description: "The name of the DBAAS resource.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "engine": schema.StringAttribute{
                Required: true,
                Description: "The database engine for the DBAAS resource (e.g., 'postgres', 'mysql').",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "backup_time": schema.StringAttribute{
                Required: true,
                Description: "The time for the daily backup in HH:MM format.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "disk_size_gb": schema.Int64Attribute{
                Required: true,
                Description: "The size of the disk in GB for the DBAAS resource.",
                PlanModifiers: []planmodifier.Int64{
                    int64planmodifier.RequiresReplace(),
                },
            },
            "password": schema.StringAttribute{
                Required: true,
                Sensitive: true,
                Description: "The password for the DBAAS resource.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "nodes_number": schema.Int32Attribute{
                Required: true,
                Description: "The number of nodes for the DBAAS resource.",
                PlanModifiers: []planmodifier.Int32{
                    int32planmodifier.RequiresReplace(),
                },
            },
            "region": schema.StringAttribute{
                Required: true,
                Description: "The region where the DBAAS resource is located.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "creation_date": schema.StringAttribute{
                Computed: true,
                Description: "The creation date of the DBAAS resource.",
            },
            "uuid": schema.StringAttribute{
           	    Computed: true,
                Description: "The UUID of the DBAAS resource.",
            },
            "last_updated": schema.StringAttribute{
               	Computed: true,
                Description: "The last updated timestamp of the DBAAS resource.",
            },
            "billing_frequency": schema.StringAttribute{
                Required: true,
                Description: "The billing frequency for the DBAAS resource (e.g., 'monthly', 'yearly').",
            },
            "billing_uuid": schema.StringAttribute{
                Computed: true,
                Description: "The UUID of the billing information for the DBAAS resource.",
            },
            "payment_url": schema.StringAttribute{
                Computed: true,
                Description: "The URL for the payment information for the DBAAS resource.",
            },
            "database_host": schema.StringAttribute{
                Computed: true,
                Description: "The host address of the database for the DBAAS resource.",
            },
            "database_port": schema.StringAttribute{
                Computed: true,
                Description: "The port number of the database for the DBAAS resource.",
            },
            "database_user": schema.StringAttribute{
                Computed: true,
                Description: "The username for the database of the DBAAS resource.",
            },
            "database_password": schema.StringAttribute{
                Computed: true,
                Sensitive: true,
                Description: "The password for the database of the DBAAS resource.",
            },
            "database_name": schema.StringAttribute{
                Computed: true,
                Description: "The name of the database for the DBAAS resource.",
            },
            "database_ssl_mode": schema.StringAttribute{
                Computed: true,
                Description: "The SSL mode for the database connection of the DBAAS resource.",
            },
            "public_uri": schema.StringAttribute{
                Computed: true,
                Description: "The public URI for accessing the DBAAS resource.",
            },
            "status": schema.StringAttribute{
                Computed: true,
                Description: "The status of the DBAAS resource.",
            },
        },
	}
}

func (r *dbaasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan common.DBAASResourceModel
    diags := req.Plan.Get(ctx, &plan)
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
        BillingFrequency: plan.BillingFrequency.ValueString(),
    }

    canCreateResponse, err := r.Client.CanCreateDBAASResource(dbaasResource)
    if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "DBAAS", err) {
	    return
    }

    response, err := r.Client.CreateDBAASResource(dbaasResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating DBAAS Resource",
            "Could not create dbaas resource, unexpected error: " + err.Error(),
        )
        return
    }

    switch res := response.(type) {
        case client.PublicCreateResponse:
            plan.BillingUUID = types.StringValue(res.BillingUUID)
            plan.PaymentURL  = types.StringValue(res.PaymentURL)
            plan.UUID        = types.StringNull()

            status, err := r.Client.GetResourceStatus(res.BillingUUID)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Error Reading DBAAS Resource Status",
                    "Could not read dbaas resource status: "+err.Error(),
                )
                return
            }

            plan.Status = types.StringValue(*status)
        case client.PrivateCreateResponse:
            plan.BillingUUID = types.StringNull()
            plan.PaymentURL  = types.StringNull()
            plan.UUID        = types.StringValue(res.ServiceUUID)
            plan.Status      = types.StringValue(res.Status)
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
}

func (r *dbaasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var model common.DBAASResourceModel
    diags := req.State.Get(ctx, &model)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    diags = model.Read(ctx, r.Client, common.RemoveResource(resp))
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
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

    err := r.Client.DeleteDBAASResource(state.UUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddWarning(
            "Error Deleting DBAAS Resource",
            "Could not delete dbaas resource, unexpected error: " + err.Error(),
        )
        return
    }
}
