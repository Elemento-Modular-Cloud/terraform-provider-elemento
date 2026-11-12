package common

import (
	"context"
	"strconv"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LoadBalancerDataSourceModel struct {
	BaseDataSourceModel
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	Port            types.Int64  `tfsdk:"port"`
	Scheme          types.String `tfsdk:"scheme"`
	DNSName         types.String `tfsdk:"dns_name"`
	ARN             types.String `tfsdk:"arn"`
	TargetGroupARN  types.String `tfsdk:"target_group_arn"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	VPCID           types.String `tfsdk:"vpc_id"`
	State           types.String `tfsdk:"state"`
}

type LoadBalancerResourceModel struct {
	LoadBalancerDataSourceModel
	ResourceExtraModel
}

func (m *LoadBalancerDataSourceModel) Read(ctx context.Context, c *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var resource *client.LoadBalancerResponse
	if m.UUID.IsNull() || m.UUID.IsUnknown() || m.UUID.ValueString() == "" {
		resources, err := client.ListResources[client.LoadBalancerResponse](c, "loadbalancer")
		if err != nil {
			diags.AddError(
				"Error Reading LoadBalancer Resources",
				"Could not list loadbalancer resources: "+err.Error(),
			)
			return true, diags
		}

		for _, res := range resources {
			if res.BillingUUID == m.BillingUUID.ValueString() {
				resource = &res
				break
			}
		}
	} else {
		res, code, err := client.GetResource[client.LoadBalancerResponse](c, m.UUID.ValueString(), "loadbalancer")
		if err != nil && code != 404 {
			diags.AddError(
				"Error Reading LoadBalancer Resource"+strconv.Itoa(code),
				"Could not read loadbalancer resource: "+err.Error(),
			)
			return true, diags
		}
		resource = res
	}

	billing_uuid := m.BillingUUID.ValueStringPointer()
	var status *string
	if billing_uuid != nil && *billing_uuid != "" {
		var err error
		status, err = c.GetResourceStatus(m.BillingUUID.ValueString())
		if err != nil {
			diags.AddError(
				"Error Reading Compute Resource Status",
				"Could not read compute resource status: "+err.Error(),
			)
			return true, diags
		}
	} else {
		status = nil
	}

	if resource == nil {
		statusDeleted := status == nil || *status == "deleted" || *status == "error" || *status == "canceled"
		onNotFound(ctx, statusDeleted)
		return !statusDeleted, diags
	}

	m.Status = types.StringPointerValue(status)
	m.Name = types.StringValue(resource.Name)
	m.Region = types.StringValue(resource.Region)
	m.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	m.UUID = types.StringValue(resource.ServiceUUID)
	m.CreatorUUID = types.StringValue(resource.CreatorUUID)
	m.CreationDate = types.StringValue(resource.CreationDate)
	m.Type = types.StringValue(resource.Type)
	m.IsPublic = types.BoolValue(resource.IsPublic)
	m.Port = types.Int64Value(resource.Port)
	m.Scheme = types.StringValue(resource.Scheme)
	m.DNSName = types.StringValue(resource.DNSName)
	m.ARN = types.StringValue(resource.ARN)
	m.TargetGroupARN = types.StringValue(resource.TargetGroupARN)
	m.SecurityGroupID = types.StringValue(resource.SecurityGroupID)
	m.VPCID = types.StringValue(resource.VPCID)
	m.State = types.StringValue(resource.State)

	// // NOTE: LoadBalancer credentials contain no useful extra information
	// credentials, err := c.GetLoadBalancerCredentials(resource.ServiceUUID)
	// if err != nil {
	// 	diags.AddWarning(
	// 		"Error Reading LoadBalancer Credentials",
	// 		"Could not read loadbalancer credentials: "+err.Error(),
	// 	)
	// 	return true, diags
	// }

	// state.NFSHost = types.StringValue(credentials.NFSHost)
	// state.NFSPath = types.StringValue(credentials.NFSPath)
	// state.NFSVersion = types.StringValue(credentials.NFSVersion)
	// state.MountCommand = types.StringValue(credentials.MountCommand)

	return true, diags
}
