package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type baseResourceModel struct {
    CreatorUUID      types.String   `tfsdk:"creator_uuid"`
    Region           types.String   `tfsdk:"region"`
    CreationDate     types.String   `tfsdk:"creation_date"`
    UUID             types.String   `tfsdk:"uuid"`
    LastUpdated      types.String   `tfsdk:"last_updated"`
    BillingFrequency types.String   `tfsdk:"billing_frequency"`
    BillingUUID      types.String   `tfsdk:"billing_uuid"`
    PaymentURL       types.String   `tfsdk:"payment_url"`
    Status           types.String   `tfsdk:"status"`
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
