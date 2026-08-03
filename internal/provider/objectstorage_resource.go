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
	_ resource.Resource = &objectStorageResource{}
	_ resource.ResourceWithConfigure = &objectStorageResource{}
)

func NewObjectStorageResource() resource.Resource {
	return &objectStorageResource{}
}

type objectStorageResource struct{
	client *client.ElementoClient
}

type objectStorageResourceModel struct {
    ActiveStorage    types.Int32    `tfsdk:"active_storage"`
    Name             types.String   `tfsdk:"name"`
    SizeTB           types.Int64    `tfsdk:"size_tb"`
    Region           types.String   `tfsdk:"region"`
    UUID             types.String   `tfsdk:"uuid"`
    Endpoint         types.String   `tfsdk:"endpoint"`
    LastUpdated      types.String   `tfsdk:"last_updated"`
    BillingFrequency types.String   `tfsdk:"billing_frequency"`
    BillingUUID      types.String   `tfsdk:"billing_uuid"`
    PaymentURL       types.String   `tfsdk:"payment_url"`
    AccountName      types.String   `tfsdk:"account_name"`
    ContainerName    types.String   `tfsdk:"container_name"`
    SasToken         types.String   `tfsdk:"sas_token"`
    AzureCliCommand  types.String   `tfsdk:"azure_cli_command"`
    Status           types.String   `tfsdk:"status"`
}

func (r *objectStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_objectstorage"
}

func (r *objectStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "name": schema.StringAttribute{
                Required: true,
            },
            "size_tb": schema.Int64Attribute{
                Required: true,
            },
            "region": schema.StringAttribute{
                Required: true,
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
            "active_storage": schema.Int32Attribute{
                Computed: true,
            },
            "endpoint": schema.StringAttribute{
                Computed: true,
            },
            "account_name": schema.StringAttribute{
                Computed: true,
            },
            "container_name": schema.StringAttribute{
                Computed: true,
            },
            "sas_token": schema.StringAttribute{
                Computed: true,
            },
            "azure_cli_command": schema.StringAttribute{
                Computed: true,
            },
            "status": schema.StringAttribute{
                Computed: true,
            },
        },
	}
}

func (r *objectStorageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *objectStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan objectStorageResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    objectStorageResource := client.ObjectStorageRequest{
        Name:             plan.Name.ValueString(),
        SizeTB:           int(plan.SizeTB.ValueInt64()),
        Region:           plan.Region.ValueString(),
        BillingFrequency: plan.BillingFrequency.ValueString(),
    }

    canCreateResponse, err := r.client.CanCreateObjectStorageResource(objectStorageResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Checking ObjectStorage Resource Creation",
            "Could not check if objectstorage resource can be created, unexpected error: " + err.Error(),
        )
        return
    }

    if !canCreateResponse.CanCreate {
        for field, reason := range canCreateResponse.Reasons {
            resp.Diagnostics.AddError(
                "ObjectStorage Resource Creation Not Allowed",
                fmt.Sprintf("Cannot create objectstorage resource due to field '%s value': %s", field, reason),
            )
        }
        return
    }

    response, err := r.client.CreateObjectStorageResource(objectStorageResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating ObjectStorage Resource",
            "Could not create objectstorage resource, unexpected error: " + err.Error(),
        )
        return
    }

    plan.BillingUUID = types.StringValue(response.BillingUUID)
    plan.PaymentURL = types.StringValue(response.PaymentURL)
    plan.UUID = types.StringNull()
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    plan.ActiveStorage = types.Int32Null()
    plan.Endpoint = types.StringNull()
    plan.AccountName = types.StringNull()
    plan.ContainerName = types.StringNull()
    plan.SasToken = types.StringNull()
    plan.AzureCliCommand = types.StringNull()

    status, err := r.client.GetResourceStatus(response.BillingUUID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading ObjectStorage Resource Status",
            "Could not read objectstorage resource status: "+err.Error(),
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

func (r *objectStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state objectStorageResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var resource *client.ObjectStorageResponse
    if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
        resources, err := r.client.ListObjectStorageResources()
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading ObjectStorage Resources",
                "Could not list objectstorage resources: "+err.Error(),
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
        res, err := r.client.GetObjectStorageResource(state.UUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading ObjectStorage Resource",
                "Could not read objectstorage resource: "+err.Error(),
            )
            return
        }
        resource = res
    }

    status, err := r.client.GetResourceStatus(state.BillingUUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading ObjectStorage Resource Status",
            "Could not read objectstorage resource status: "+err.Error(),
        )
        return
    }

    if resource == nil && (*status == "deleted" || *status == "error") {
        resp.State.RemoveResource(ctx)
        return
    }

    state.Status        = types.StringValue(*status)
    state.Name          = types.StringValue(resource.Name)
    state.ActiveStorage = types.Int32Value(int32(resource.ActiveStorage))
    state.Endpoint      = types.StringValue(resource.Endpoint)
    state.Region        = types.StringValue(resource.Region)
    state.LastUpdated   = types.StringValue(time.Now().Format(time.RFC850))
    state.UUID          = types.StringValue(resource.UUID)
    state.PaymentURL    = types.StringNull()

    credentials, err := r.client.GetObjectStorageCredentials(resource.UUID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading ObjectStorage Credentials",
            "Could not read objectstorage credentials: "+err.Error(),
        )
        return
    }

    state.AccountName = types.StringValue(credentials.AccountName)
    state.ContainerName = types.StringValue(credentials.ContainerName)
    state.SasToken = types.StringValue(credentials.SasToken)
    state.AzureCliCommand = types.StringValue(credentials.AzureCliCommand)

    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *objectStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *objectStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state objectStorageResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.client.DeleteObjectStorageResource(state.UUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddWarning(
            "Error Deleting ObjectStorage Resource",
            "Could not delete objectstorage resource, unexpected error: " + err.Error(),
        )
        return
    }
}
