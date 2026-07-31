package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	client *client.ElementoClient
}

type dbaasResourceModel struct {
    CreatorUUID      types.String   `tfsdk:"creator_uuid"`
    Name             types.String   `tfsdk:"name"`
    Engine           types.String   `tfsdk:"engine"`
    BackupTime       types.String   `tfsdk:"backup_time"`
    DiskSizeGB       types.Int64    `tfsdk:"disk_size_gb"`
    Region           types.String   `tfsdk:"region"`
    Password         types.String   `tfsdk:"password"`
    NodesNumber      types.Int32    `tfsdk:"nodes_number"`
    CreationDate     types.String   `tfsdk:"creation_date"`
    UUID             types.String   `tfsdk:"uuid"`
    LastUpdated      types.String   `tfsdk:"last_updated"`
    BillingFrequency types.String   `tfsdk:"billing_frequency"`
    BillingUUID      types.String   `tfsdk:"billing_uuid"`
    PaymentURL       types.String   `tfsdk:"payment_url"`
    DatabaseHost     types.String   `tfsdk:"database_host"`
    DatabasePort     types.String   `tfsdk:"database_port"`
    DatabaseUser     types.String   `tfsdk:"database_user"`
    DatabasePassword types.String   `tfsdk:"database_password"`
    DatabaseName     types.String   `tfsdk:"database_name"`
    DatabaseSSLMode  types.String   `tfsdk:"database_ssl_mode"`
    PublicURI        types.String   `tfsdk:"public_uri"`
    Status           types.String   `tfsdk:"status"`
}

func (r *dbaasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbaas"
}

func (r *dbaasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "creator_uuid": schema.StringAttribute{
                Computed: true,
            },
            "name": schema.StringAttribute{
                Required: true,
            },
            "engine": schema.StringAttribute{
                Required: true,
            },
            "backup_time": schema.StringAttribute{
                Required: true,
            },
            "disk_size_gb": schema.Int64Attribute{
                Required: true,
            },
            "password": schema.StringAttribute{
                Required: true,
                Sensitive: true,
            },
            "nodes_number": schema.Int32Attribute{
                Required: true,
            },
            "region": schema.StringAttribute{
                Required: true,
            },
            "creation_date": schema.StringAttribute{
                Computed: true,
            },
            "uuid": schema.StringAttribute{
           	    Computed: true,
            },
            "last_updated": schema.StringAttribute{
               	Computed: true,
            },
            "billing_frequency": schema.StringAttribute{
                Required: true,
            },
            "billing_uuid": schema.StringAttribute{
                Computed: true,
            },
            "payment_url": schema.StringAttribute{
                Computed: true,
            },
            "database_host": schema.StringAttribute{
                Computed: true,
            },
            "database_port": schema.StringAttribute{
                Computed: true,
            },
            "database_user": schema.StringAttribute{
                Computed: true,
            },
            "database_password": schema.StringAttribute{
                Computed: true,
                Sensitive: true,
            },
            "database_name": schema.StringAttribute{
                Computed: true,
            },
            "database_ssl_mode": schema.StringAttribute{
                Computed: true,
            },
            "public_uri": schema.StringAttribute{
                Computed: true,
            },
            "status": schema.StringAttribute{
                Computed: true,
            },
        },
	}
}

func (r *dbaasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dbaasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan dbaasResourceModel
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

    canCreateResponse, err := r.client.CanCreateDBAASResource(dbaasResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Checking DBAAS Resource Creation",
            "Could not check if dbaas resource can be created, unexpected error: " + err.Error(),
        )
        return
    }

    if !canCreateResponse.CanCreate {
        for field, reason := range canCreateResponse.Reasons {
            resp.Diagnostics.AddError(
                "DBAAS Resource Creation Not Allowed",
                fmt.Sprintf("Cannot create dbaas resource for field '%s': %s", field, reason),
            )
        }
        return
    }

    response, err := r.client.CreateDBAASResource(dbaasResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating DBAAS Resource",
            "Could not create dbaas resource, unexpected error: " + err.Error(),
        )
        return
    }

    plan.BillingUUID = types.StringValue(response.BillingUUID)
    plan.PaymentURL = types.StringValue(response.PaymentURL)
    plan.UUID = types.StringNull()
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

    status, err := r.client.GetResourceStatus(response.BillingUUID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading DBAAS Resource Status",
            "Could not read dbaas resource status: "+err.Error(),
        )
        return
    }

    plan.Status = types.StringValue(*status)

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *dbaasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state dbaasResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var resource *client.DBAASResponse
    if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
        resources, err := r.client.ListDBAASResources()
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading DBAAS Resources",
                "Could not list dbaas resources: "+err.Error(),
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
        res, err := r.client.GetDBAASResource(state.UUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading DBAAS Resource",
                "Could not read dbaas resource: "+err.Error(),
            )
            return
        }
        resource = res
    }

    status, err := r.client.GetResourceStatus(state.BillingUUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading DBAAS Resource Status",
            "Could not read dbaas resource status: "+err.Error(),
        )
        return
    }

    if resource == nil && (*status == "deleted" || *status == "error") {
        resp.State.RemoveResource(ctx)
        return
    }

    state.Status       = types.StringValue(*status)
    state.Name         = types.StringValue(resource.Name)
    state.Region       = types.StringValue(resource.Region)
    state.LastUpdated  = types.StringValue(time.Now().Format(time.RFC850))
    state.UUID         = types.StringValue(resource.UUID)
    state.CreatorUUID  = types.StringValue(resource.CreatorUUID)
    state.CreationDate = types.StringValue(resource.CreationDate)
    state.BackupTime   = types.StringValue(resource.BackupTime)
    state.DiskSizeGB   = types.Int64Value(resource.DiskSizeGB)
    state.NodesNumber  = types.Int32Value(resource.NodesNumber)
    state.Engine       = types.StringValue(resource.Engine)
    state.PaymentURL   = types.StringNull()

    credentials, err := r.client.GetDBAASCredentials(resource.UUID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading DBAAS Credentials",
            "Could not read dbaas credentials: "+err.Error(),
        )
        return
    }

    state.DatabaseHost = types.StringValue(credentials.DatabaseHost)
    state.DatabasePort = types.StringValue(credentials.DatabasePort)
    state.DatabaseUser = types.StringValue(credentials.DatabaseUser)
    state.DatabasePassword = types.StringValue(credentials.DatabasePassword)
    state.DatabaseName = types.StringValue(credentials.DatabaseName)
    state.DatabaseSSLMode = types.StringValue(credentials.DatabaseSSLMode)
    state.PublicURI = types.StringValue(credentials.PublicURI)

    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *dbaasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *dbaasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state dbaasResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.client.DeleteDBAASResource(state.UUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddWarning(
            "Error Deleting DBAAS Resource",
            "Could not delete dbaas resource, unexpected error: " + err.Error(),
        )
        return
    }
}
