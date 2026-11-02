package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
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
    ValidateCanCreate(canCreateResponse, &diags, "DBAAS", err)

    response, err := r.client.CreateDBAASResource(dbaasResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating DBAAS Resource",
            "Could not create dbaas resource, unexpected error: " + err.Error(),
        )
        return
    }

    switch res := response.(type) {
        case *client.PublicCreateResponse:
            plan.BillingUUID = types.StringValue(res.BillingUUID)
            plan.PaymentURL  = types.StringValue(res.PaymentURL)
            plan.UUID        = types.StringNull()

            status, err := r.client.GetResourceStatus(res.BillingUUID)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Error Reading DBAAS Resource Status",
                    "Could not read dbaas resource status: "+err.Error(),
                )
                return
            }

            plan.Status = types.StringValue(*status)
        case *client.PrivateCreateResponse:
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
        res, code, err := r.client.GetDBAASResource(state.UUID.ValueString())
        if err != nil && code != 404 {
            resp.Diagnostics.AddError(
                "Error Reading DBAAS Resource",
                "Could not read dbaas resource: "+err.Error(),
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

    state.Status       = types.StringPointerValue(status)
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
