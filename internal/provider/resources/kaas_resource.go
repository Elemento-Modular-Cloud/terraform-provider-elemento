package resources

import (
	"context"
	"terraform-provider-elemento/internal/client"
	"terraform-provider-elemento/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &kaasResource{}
	_ resource.ResourceWithConfigure = &kaasResource{}
	_ resource.ResourceWithImportState = &kaasResource{}
)

func NewKAASResource() resource.Resource {
	return &kaasResource{}
}

type kaasResource struct{
	common.BaseResource
}

func (r *kaasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kaas"
}

func (r *kaasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := common.BaseSchema()

    attributes["name"] = schema.StringAttribute{
        Required: true,
        Description: "The name of the KAAS resource.",
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.RequiresReplace(),
        },
    }
    attributes["node_pools"] = schema.ListNestedAttribute{
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
    }
    attributes["nodes_subnet"] = schema.SingleNestedAttribute{
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
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
    }
    attributes["update_policy"] = schema.StringAttribute{
        Required: true,
        Description: "The update policy for the KAAS resource (e.g., 'ALWAYS').",
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.RequiresReplace(),
        },
    }
    attributes["kubernetes_version"] = schema.StringAttribute{
        Required: true,
        Description: "The Kubernetes version for the KAAS resource.",
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.RequiresReplace(),
        },
    }
	attributes["network_cidr"] = schema.StringAttribute{
		Computed:    true,
		Description: "Cluster network CIDR.",
	}

    attributes["control_plane_ha_enabled"] = schema.BoolAttribute{
        Required: true,
        Description: "Whether high availability is enabled for the control plane.",
        PlanModifiers: []planmodifier.Bool{
            boolplanmodifier.RequiresReplace(),
        },
    }
    attributes["cluster_name"] = schema.StringAttribute{
        Computed: true,
        Description: "The name of the Kubernetes cluster associated with the KAAS resource.",
    }
    attributes["cluster_status"] = schema.StringAttribute{
        Computed: true,
        Description: "The status of the Kubernetes cluster associated with the KAAS resource.",
    }
    attributes["endpoint"] = schema.StringAttribute{
        Computed: true,
        Description: "The endpoint for the KAAS resource.",
    }
    attributes["kube_config"] = schema.StringAttribute{
        Computed: true,
        Description: "The kubeconfig for the Kubernetes cluster associated with the KAAS resource.",
        Sensitive: true,
    }

    resp.Schema = schema.Schema{
        Attributes: attributes,
    }
}

func (r *kaasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan common.KAASResourceModel
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
        BillingFrequency: plan.BillingFrequency.ValueStringPointer(),
        NodePools:        nodePools,
        NodesSubnet: client.KaasNodeSubnet{
            DHCPEnabled: plan.NodesSubnet.DHCPEnabled.ValueBool(),
            CIDR:        plan.NodesSubnet.CIDR.ValueString(),
        },
        UpdatePolicy:          plan.UpdatePolicy.ValueString(),
        KubernetesVersion:     plan.Version.ValueString(),
        ControlPlaneHAEnabled: plan.ControlPlaneHAEnabled.ValueBool(),
    }

    canCreateResponse, err := r.Client.CanCreateKAASResource(ctx, kaasResource)
    if !common.ValidateCanCreate(canCreateResponse, &resp.Diagnostics, "KAAS", err) {
	    return
    }

    response, err := r.Client.CreateKAASResource(ctx, kaasResource)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating KAAS Resource",
            "Could not create kaas resource, unexpected error: " + err.Error(),
        )
        return
    }

    switch res := response.(type) {
        case client.PublicCreateResponse:
            plan.BillingUUID = types.StringValue(res.BillingUUID)
            plan.PaymentURL  = types.StringValue(res.PaymentURL)
            plan.UUID        = types.StringNull()

			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)

            status, err := r.Client.GetResourceStatus(ctx, res.BillingUUID)
            if err != nil {
                resp.Diagnostics.AddWarning(
                    "Error Reading KAAS Resource Status",
                    "Could not read kaas resource status: "+err.Error(),
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
    plan.CreatorUUID = types.StringNull()
    plan.CreationDate = types.StringNull()

    plan.ClusterName   = types.StringNull()
    plan.ClusterStatus = types.StringNull()
    plan.Endpoint      = types.StringNull()
    plan.KubeConfig    = types.StringNull()

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *kaasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var model common.KAASResourceModel
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

func (r *kaasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // TODO: Implement update if the API supports it. For now, we mark everything as needing replacement.
}

func (r *kaasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state common.KAASResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    if state.UUID.IsNull() {
	    resp.Diagnostics.AddError(
		    "Error Deleting KAAS Resource",
		    "Could not delete kaas resource: missing resource uuid is null.",
	    )
		return
    }

    statusCode, err := r.Client.DeleteKAASResource(ctx, state.UUID.ValueString())
    if err != nil {
		if statusCode == 404 {
			resp.Diagnostics.AddWarning(
				"Error Deleting KAAS Resource",
				"Could not delete kaas resource: resource not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting KAAS Resource",
				"Could not delete kaas resource, unexpected error: " + err.Error(),
			)
		}
		return
    }
}

func (r *kaasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
