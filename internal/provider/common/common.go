package common

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BaseDataSourceModel struct {
	CreatorUUID      types.String `tfsdk:"creator_uuid"`
	Region           types.String `tfsdk:"region"`
	CreationDate     types.String `tfsdk:"creation_date"`
	UUID             types.String `tfsdk:"uuid"`
	LastUpdated      types.String `tfsdk:"last_updated"`
	BillingFrequency types.String `tfsdk:"billing_frequency"`
	BillingUUID      types.String `tfsdk:"billing_uuid"`
	Status           types.String `tfsdk:"status"`
}

type ResourceExtraModel struct {
	PaymentURL types.String `tfsdk:"payment_url"`
}

func BaseSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"creator_uuid": rschema.StringAttribute{
			Computed:    true,
			Description: "The UUID of the creator of the resource.",
		},
		"region": rschema.StringAttribute{
			Required:    true,
			Description: "The region where the resource is located.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"creation_date": rschema.StringAttribute{
			Computed:    true,
			Description: "The creation date of the resource.",
		},
		"uuid": rschema.StringAttribute{
			Computed:    true,
			Description: "The UUID of the resource.",
		},
		"last_updated": rschema.StringAttribute{
			Computed:    true,
			Description: "The last updated timestamp of the resource.",
		},
		"billing_frequency": rschema.StringAttribute{
			Optional:    true,
			Description: "The billing frequency for the resource (e.g., 'monthly', 'yearly'). Required for public mesons.",
		},
		"billing_uuid": rschema.StringAttribute{
			Computed:    true,
			Description: "The UUID of the billing information for the resource.",
		},
		"payment_url": rschema.StringAttribute{
			Computed:    true,
			Description: "The URL for the payment information for the resource.",
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: "The status of the resource.",
		},
	}
}

func identifierExactlyOneOf(paths ...string) []validator.String {
	exprs := make([]path.Expression, 0, len(paths))
	for _, p := range paths {
		exprs = append(exprs, path.MatchRoot(p))
	}
	return []validator.String{stringvalidator.ExactlyOneOf(exprs...)}
}

func BaseDataSourceSchema() map[string]dschema.Attribute {
	ids := identifierExactlyOneOf("uuid", "billing_uuid")

	return map[string]dschema.Attribute{
		"uuid": dschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Service UUID of the datasource. Exactly one of uuid or billing_uuidmust be set.",
			Validators:  ids,
		},
		"billing_uuid": dschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Portal billing UUID of the datasource.",
			Validators:  ids,
		},
		"creator_uuid": dschema.StringAttribute{
			Computed:    true,
			Description: "The UUID of the creator of the datasource.",
		},
		"region": dschema.StringAttribute{
			Computed:    true,
			Description: "The region where the datasource is located.",
		},
		"creation_date": dschema.StringAttribute{
			Computed:    true,
			Description: "The creation date of the datasource.",
		},
		"last_updated": dschema.StringAttribute{
			Computed:    true,
			Description: "The last updated timestamp of the datasource.",
		},
		"billing_frequency": dschema.StringAttribute{
			Computed:    true,
			Description: "The billing frequency for the datasource (e.g., 'monthly', 'yearly').",
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: "The status of the resource.",
		},
	}
}


type BaseResource struct {
	Client *client.ElementoClient
}

type BaseDataSource struct {
	Client *client.ElementoClient
}

// TODO: use in older resources
func (r *BaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.Client = client
}

func (r *BaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	r.Client = client
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

func RemoveResource(resp *resource.ReadResponse) func(context.Context) {
	return func(ctx context.Context) {
		resp.State.RemoveResource(ctx)
	}
}

func AddNotFoundError(diags diag.Diagnostics, typeName string) func(context.Context) {
	return func(ctx context.Context) {
		diags.AddError(
			fmt.Sprintf("No matching %s Found", typeName),
			fmt.Sprintf(
				"No running %s matched the given identifier. Check uuid or billing_uuid and the provider target.",
				strings.ToLower(typeName),
			),
		)
	}
}
