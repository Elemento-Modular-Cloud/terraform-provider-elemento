package common

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"terraform-provider-elemento/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
	PaymentURL types.String   `tfsdk:"payment_url"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

func BaseSchema(ctx context.Context) map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"creator_uuid": rschema.StringAttribute{
			Computed:    true,
			Description: "UUID of the account that created the resource.",
		},
		"region": rschema.StringAttribute{
			Required:    true,
			Description: "Region to deploy in. Changing it forces a new resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"creation_date": rschema.StringAttribute{
			Computed:    true,
			Description: "Creation timestamp.",
		},
		"uuid": rschema.StringAttribute{
			Computed:    true,
			Description: "Service UUID.",
		},
		"last_updated": rschema.StringAttribute{
			Computed:    true,
			Description: "Timestamp of the last update.",
		},
		"billing_frequency": rschema.StringAttribute{
			Optional:    true,
			Description: "Billing period, e.g. `day`, `month` or `year`. Changing it forces a new resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"billing_uuid": rschema.StringAttribute{
			Computed:    true,
			Description: "Portal billing UUID.",
		},
		"payment_url": rschema.StringAttribute{
			Computed:    true,
			Description: "URL where the payment for the resource is completed.",
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: "Current status of the resource.",
		},
		"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
			Create: true,
		}),
	}
}

func identifierExactlyOneOf(paths ...string) validator.String {
	exprs := make([]path.Expression, 0, len(paths))
	for _, p := range paths {
		exprs = append(exprs, path.MatchRoot(p))
	}
	return stringvalidator.ExactlyOneOf(exprs...)
}

func BaseDataSourceSchema() map[string]dschema.Attribute {
	ids := identifierExactlyOneOf("uuid", "billing_uuid")
	uuid := stringvalidator.RegexMatches(
		regexp.MustCompile("^(?i)([0-9a-f]{32}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$"),
		"value must be a valid UUID.",
	)

	return map[string]dschema.Attribute{
		"uuid": dschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Service UUID. Set exactly one of `uuid` or `billing_uuid`.",
			Validators:  []validator.String{ids, uuid},
		},
		"billing_uuid": dschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Portal billing UUID. Set exactly one of `uuid` or `billing_uuid`.",
			Validators:  []validator.String{ids, uuid},
		},
		"creator_uuid": dschema.StringAttribute{
			Computed:    true,
			Description: "UUID of the account that created the resource.",
		},
		"region": dschema.StringAttribute{
			Computed:    true,
			Description: "Region the resource is in.",
		},
		"creation_date": dschema.StringAttribute{
			Computed:    true,
			Description: "Creation timestamp.",
		},
		"last_updated": dschema.StringAttribute{
			Computed:    true,
			Description: "Timestamp of the last update.",
		},
		"billing_frequency": dschema.StringAttribute{
			Computed:    true,
			Description: "Billing period.",
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: "Current status of the resource.",
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

type OnNotFound func(ctx context.Context, statusDeleted bool)

func RemoveResource(resp *resource.ReadResponse) OnNotFound {
	return func(ctx context.Context, statusDeleted bool) {
		if statusDeleted {
			resp.State.RemoveResource(ctx)
		}
	}
}

func AddNotFoundError(diags *diag.Diagnostics, typeName string) OnNotFound {
	return func(ctx context.Context, _ bool) {
		diags.AddError(
			fmt.Sprintf("No matching %s Found", typeName),
			fmt.Sprintf(
				"No running %s matched the given identifier. Check uuid or billing_uuid and the provider target.",
				strings.ToLower(typeName),
			),
		)
	}
}

const DefaultCreateTimeout = 20 * time.Minute
const DefaultDeleteTimeout = 20 * time.Minute

type HasRead interface {
	Read(ctx context.Context, client *client.ElementoClient, onNotFound OnNotFound) (bool, diag.Diagnostics)
}

func GetCreateStateConf(ctx context.Context, client *client.ElementoClient, timeout time.Duration, plan HasRead) *retry.StateChangeConf {
    	return &retry.StateChangeConf{
     		Pending: []string{"CREATING"},
       		Target: []string{"READY"},
         	Delay: 5 * time.Second,
          	MinTimeout: 3 * time.Second,
           	Timeout: timeout,
            Refresh: func() (any, string, error) {
            	found, diags := plan.Read(ctx, client, func(_ context.Context, _ bool) {})
             	if diags.HasError() {
             		return nil, "", fmt.Errorf("%s: %s", diags.Errors()[0].Summary(), diags.Errors()[0].Detail())
             	}
              	if !found {
               		return &plan, "CREATING", nil
                }
                return &plan, "READY", nil
            },
     	}

}
