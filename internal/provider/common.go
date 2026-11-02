package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type baseResourceModel struct {
	CreatorUUID      types.String `tfsdk:"creator_uuid"`
	Region           types.String `tfsdk:"region"`
	CreationDate     types.String `tfsdk:"creation_date"`
	UUID             types.String `tfsdk:"uuid"`
	LastUpdated      types.String `tfsdk:"last_updated"`
	BillingFrequency types.String `tfsdk:"billing_frequency"`
	BillingUUID      types.String `tfsdk:"billing_uuid"`
	PaymentURL       types.String `tfsdk:"payment_url"`
	Status           types.String `tfsdk:"status"`
}

func baseSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"creator_uuid": schema.StringAttribute{
			Computed:    true,
			Description: "The UUID of the creator of the resource.",
		},
		"region": schema.StringAttribute{
			Required:    true,
			Description: "The region where the resource is located.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"creation_date": schema.StringAttribute{
			Computed:    true,
			Description: "The creation date of the resource.",
		},
		"uuid": schema.StringAttribute{
			Computed:    true,
			Description: "The UUID of the resource.",
		},
		"last_updated": schema.StringAttribute{
			Computed:    true,
			Description: "The last updated timestamp of the resource.",
		},
		"billing_frequency": schema.StringAttribute{
			Optional:    true,
			Description: "The billing frequency for the resource (e.g., 'monthly', 'yearly'). Required for public mesons.",
		},
		"billing_uuid": schema.StringAttribute{
			Computed:    true,
			Description: "The UUID of the billing information for the resource.",
		},
		"payment_url": schema.StringAttribute{
			Computed:    true,
			Description: "The URL for the payment information for the resource.",
		},
		"status": schema.StringAttribute{
			Computed:    true,
			Description: "The status of the resource.",
		},
	}
}

type baseResource struct {
	client *client.ElementoClient
}

// TODO: use in older resources
func (r *baseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func ValidateCanCreate(check *client.CanCreateResponse, diags *diag.Diagnostics, resourceName string, err error) bool {
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error Checking %s Resource Creation", resourceName),
			fmt.Sprintf(
				"Could not check if %s resource can be created, unexpected error: %s",
				strings.ToLower(resourceName),
				err.Error(),
			),
		)
		return false
	}

	if !check.CanCreate {
		if len(check.Reasons) == 0 {
			diags.AddError(
				fmt.Sprintf("%s Resource Creation Not Allowed", resourceName),
				fmt.Sprintf(
					"Cannot create %s resource due to an unknown reason. Please contact the devs or try again later.",
					strings.ToLower(resourceName),
				),
			)
		}

		for field, reason := range check.Reasons {
			diags.AddError(
				fmt.Sprintf("%s Resource Creation Not Allowed", resourceName),
				fmt.Sprintf(
					"Cannot create %s resource due to field '%s' value: %s",
					strings.ToLower(resourceName),
					field,
					reason,
				),
			)
		}
		return false
	}

	return true
}
