package provider

import (
	"context"
	"fmt"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &kaasResource{}
	_ resource.ResourceWithConfigure = &kaasResource{}
)

func NewKAASResource() resource.Resource {
	return &kaasResource{}
}

type kaasResource struct{
	client *client.ElementoClient
}

type kaasResourceModel struct {
    CreatorUUID           types.String      `tfsdk:"creator_uuid"`
    Name                  types.String      `tfsdk:"name"`
    Region                types.String      `tfsdk:"region"`
    UUID                  types.String      `tfsdk:"uuid"`
    LastUpdated           types.String      `tfsdk:"last_updated"`
    BillingFrequency      types.String      `tfsdk:"billing_frequency"`
    BillingUUID           types.String      `tfsdk:"billing_uuid"`
    PaymentURL            types.String      `tfsdk:"payment_url"`
    NodePools             []nodePoolModel   `tfsdk:"node_pools"`
    NodesSubnet           nodeSubnetModel   `tfsdk:"nodes_subnet"`
    UpdatePolicy          types.String      `tfsdk:"update_policy"`
    KubernetesVersion     types.String      `tfsdk:"kubernetes_version"`
    ControlPlaneHAEnabled types.Bool        `tfsdk:"control_plane_ha_enabled"`
    ClusterName           types.String      `tfsdk:"cluster_name"`
    ClusterStatus         types.String      `tfsdk:"cluster_status"`
    Endpoint              types.String      `tfsdk:"endpoint"`
    KubeConfig            types.String      `tfsdk:"kube_config"`
    Status                types.String      `tfsdk:"status"`
}

type nodePoolModel struct {
    Name             types.String   `tfsdk:"name"`
    AutoScale        types.Bool     `tfsdk:"auto_scale"`
    MinNodes         types.Int32    `tfsdk:"min_nodes"`
    MaxNodes         types.Int32    `tfsdk:"max_nodes"`
    SharedStorageGB  types.Int64    `tfsdk:"shared_storage_gb"`
    Specs            nodePoolSpecsModel `tfsdk:"specs"`
}

type nodePoolSpecsModel struct {
    CPUCores         types.Int32    `tfsdk:"cpu_cores"`
    MemoryGB         types.Int32    `tfsdk:"memory_gb"`
    DiskSizeGB       types.Int32    `tfsdk:"disk_size_gb"`
    GPUs             types.Int32    `tfsdk:"gpus"`
}

type nodeSubnetModel struct {
    DHCPEnabled      types.Bool     `tfsdk:"dhcp_enabled"`
    CIDR             types.String   `tfsdk:"cidr"`
}

func (r *kaasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kaas"
}

func (r *kaasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "creator_uuid": schema.StringAttribute{
                Computed: true,
                Description: "The UUID of the user who created the KAAS resource.",
            },
            "name": schema.StringAttribute{
                Required: true,
                Description: "The name of the KAAS resource.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "region": schema.StringAttribute{
                Required: true,
                Description: "The region where the KAAS resource will be created.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "uuid": schema.StringAttribute{
           	    Computed: true,
                Description: "The unique identifier of the KAAS resource.",
            },
            "last_updated": schema.StringAttribute{
               	Computed: true,
                Description: "The timestamp of the last update to the KAAS resource.",
            },
            "billing_frequency": schema.StringAttribute{
                Required: true,
                Description: "The billing frequency for the KAAS resource (e.g., 'month', 'year').",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "billing_uuid": schema.StringAttribute{
                Computed: true,
                Description: "The billing UUID associated with the KAAS resource.",
            },
            "payment_url": schema.StringAttribute{
                Computed: true,
                Description: "The payment URL for the KAAS resource, assigned by the Elemento API after creation.",
            },
            "node_pools": schema.ListNestedAttribute{
                Required: true,
                Description: "A list of node pools for the KAAS resource.",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "name": schema.StringAttribute{
                            Required: true,
                            Description: "The name of the node pool.",
                        },
                        "auto_scale": schema.BoolAttribute{
                            Required: true,
                            Description: "Whether auto-scaling is enabled for the node pool.",
                        },
                        "min_nodes": schema.Int32Attribute{
                            Required: true,
                            Description: "The minimum number of nodes in the node pool.",
                        },
                        "max_nodes": schema.Int32Attribute{
                            Required: true,
                            Description: "The maximum number of nodes in the node pool.",
                        },
                        "shared_storage_gb": schema.Int64Attribute{
                            Required: true,
                            Description: "The amount of shared storage in GB for the node pool.",
                        },
                        "specs": schema.SingleNestedAttribute{
                            Required: true,
                            Attributes: map[string]schema.Attribute{
                                "cpu_cores": schema.Int32Attribute{
                                    Required: true,
                                    Description: "The number of CPU cores for the node pool.",
                                },
                                "memory_gb": schema.Int32Attribute{
                                    Required: true,
                                    Description: "The amount of memory in GB for the node pool.",
                                },
                                "disk_size_gb": schema.Int32Attribute{
                                    Required: true,
                                    Description: "The size of the disk in GB for the node pool.",
                                },
                                "gpus": schema.Int32Attribute{
                                    Required: true,
                                    Description: "The number of GPUs for the node pool.",
                                },
                            },
                        },
                    },
                },
                PlanModifiers: []planmodifier.List{
                    listplanmodifier.RequiresReplace(),
                },
            },
            "nodes_subnet": schema.SingleNestedAttribute{
                Required: true,
                Attributes: map[string]schema.Attribute{
                    "dhcp_enabled": schema.BoolAttribute{
                        Required: true,
                        Description: "Whether DHCP is enabled for the nodes subnet.",
                    },
                    "cidr": schema.StringAttribute{
                        Required: true,
                        Description: "The CIDR block for the nodes subnet.",
                    },
                },
            },
            "update_policy": schema.StringAttribute{
                Required: true,
                Description: "The update policy for the KAAS resource (e.g., 'ALWAYS').",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "kubernetes_version": schema.StringAttribute{
                Required: true,
                Description: "The Kubernetes version for the KAAS resource.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "control_plane_ha_enabled": schema.BoolAttribute{
                Required: true,
                Description: "Whether high availability is enabled for the control plane.",
                PlanModifiers: []planmodifier.Bool{
                    boolplanmodifier.RequiresReplace(),
                },
            },
            "cluster_name": schema.StringAttribute{
                Computed: true,
                Description: "The name of the Kubernetes cluster associated with the KAAS resource.",
            },
            "cluster_status": schema.StringAttribute{
                Computed: true,
                Description: "The status of the Kubernetes cluster associated with the KAAS resource.",
            },
            "endpoint": schema.StringAttribute{
                Computed: true,
                Description: "The endpoint for the KAAS resource.",
            },
            "kube_config": schema.StringAttribute{
                Computed: true,
                Description: "The kubeconfig for the Kubernetes cluster associated with the KAAS resource.",
            },
            "status": schema.StringAttribute{
                Computed: true,
                Description: "The status of the KAAS resource.",
            },
        },
	}
}

func (r *kaasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *kaasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan kaasResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var nodePools []client.KAASNodePool
    for _, np := range plan.NodePools {
        nodePools = append(nodePools, client.KAASNodePool{
            Name:            np.Name.ValueString(),
            AutoScale:       np.AutoScale.ValueBool(),
            MinNodes:        np.MinNodes.ValueInt32(),
            MaxNodes:        np.MaxNodes.ValueInt32(),
            SharedStorageGB: np.SharedStorageGB.ValueInt64(),
            Specs: client.KAASNodePoolSpecs{
                CPUCores:   int(np.Specs.CPUCores.ValueInt32()),
                MemoryGB:   int(np.Specs.MemoryGB.ValueInt32()),
                DiskSizeGB: int(np.Specs.DiskSizeGB.ValueInt32()),
                GPUs:       int(np.Specs.GPUs.ValueInt32()),
            },
        })
    }

    kaasResource := client.KAASRequest{
        Name:             plan.Name.ValueString(),
        Region:           plan.Region.ValueString(),
        BillingFrequency: plan.BillingFrequency.ValueString(),
        NodePools:        nodePools,
        NodesSubnet: client.KaasNodeSubnet{
            DHCPEnabled: plan.NodesSubnet.DHCPEnabled.ValueBool(),
            CIDR:        plan.NodesSubnet.CIDR.ValueString(),
        },
        UpdatePolicy:          plan.UpdatePolicy.ValueString(),
        KubernetesVersion:     plan.KubernetesVersion.ValueString(),
        ControlPlaneHAEnabled: plan.ControlPlaneHAEnabled.ValueBool(),
    }

    canCreateResponse, err := r.client.CanCreateKAASResource(kaasResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Checking KAAS Resource Creation",
            "Could not check if kaas resource can be created, unexpected error: " + err.Error(),
        )
        return
    }

    if !canCreateResponse.CanCreate {
        for field, reason := range canCreateResponse.Reasons {
            resp.Diagnostics.AddError(
                "KAAS Resource Creation Not Allowed",
                fmt.Sprintf("Cannot create kaas resource due to field '%s' value: %s", field, reason),
            )
        }
        return
    }

    response, err := r.client.CreateKAASResource(kaasResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating KAAS Resource",
            "Could not create kaas resource, unexpected error: " + err.Error(),
        )
        return
    }

    plan.BillingUUID = types.StringValue(response.BillingUUID)
    plan.PaymentURL  = types.StringValue(response.PaymentURL)
    plan.UUID        = types.StringNull()
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    plan.CreatorUUID = types.StringNull()

    plan.ClusterName   = types.StringNull()
    plan.ClusterStatus = types.StringNull()
    plan.Endpoint      = types.StringNull()
    plan.KubeConfig    = types.StringNull()

    status, err := r.client.GetResourceStatus(response.BillingUUID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading KAAS Resource Status",
            "Could not read kaas resource status: "+err.Error(),
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

func (r *kaasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state kaasResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var resource *client.KAASResponse
    if state.UUID.IsNull() || state.UUID.IsUnknown() || state.UUID.ValueString() == "" {
        resources, err := r.client.ListKAASResources()
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading KAAS Resources",
                "Could not list kaas resources: "+err.Error(),
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
        res, err := r.client.GetKAASResource(state.UUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error Reading KAAS Resource",
                "Could not read kaas resource: "+err.Error(),
            )
            return
        }
        resource = res
    }

    status, err := r.client.GetResourceStatus(state.BillingUUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading KAAS Resource Status",
            "Could not read kaas resource status: "+err.Error(),
        )
        return
    }

    if resource == nil && (*status == "deleted" || *status == "error") {
        resp.State.RemoveResource(ctx)
        return
    }

    state.Status      = types.StringValue(*status)
    state.Name        = types.StringValue(resource.Name)
    state.Region      = types.StringValue(resource.Region)
    state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    state.UUID        = types.StringValue(resource.UUID)
    state.CreatorUUID = types.StringValue(resource.CreatorUUID)
    state.PaymentURL  = types.StringNull()

    state.ClusterName   = types.StringValue(resource.ClusterName)
    state.ClusterStatus = types.StringValue(resource.Status)
    state.Endpoint      = types.StringValue(resource.Endpoint)
    state.KubeConfig    = types.StringValue(resource.KubeConfig)

    // credentials, err := r.client.GetKAASCredentials(resource.UUID)
    // if err != nil {
    //     resp.Diagnostics.AddError(
    //         "Error Reading KAAS Credentials",
    //         "Could not read kaas credentials: "+err.Error(),
    //     )
    //     return
    // }

    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *kaasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *kaasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state kaasResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.client.DeleteKAASResource(state.UUID.ValueString())
    if err != nil {
        resp.Diagnostics.AddWarning(
            "Error Deleting KAAS Resource",
            "Could not delete kaas resource, unexpected error: " + err.Error(),
        )
        return
    }
}
