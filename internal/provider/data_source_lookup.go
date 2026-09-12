package provider

import (
	"fmt"

	"terraform-provider-elemento/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func configureDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *client.ElementoClient {
	if req.ProviderData == nil {
		return nil
	}

	c, ok := req.ProviderData.(*client.ElementoClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *ElementoClient, got %T. Please report this issue to the provider developer.", req.ProviderData),
		)
		return nil
	}

	return c
}

func optionalString(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func identifierExactlyOneOf(paths ...string) []validator.String {
	exprs := make([]path.Expression, 0, len(paths))
	for _, p := range paths {
		exprs = append(exprs, path.MatchRoot(p))
	}
	return []validator.String{stringvalidator.ExactlyOneOf(exprs...)}
}

func requireOneMatch[T any](matches []T, typeName string) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch len(matches) {
	case 0:
		diags.AddError(
			"No matching "+typeName,
			"No running "+typeName+" matched the given identifier. Check uuid, billing_uuid, or name, and that the provider host targets the intended cloud.",
		)
		return nil, diags
	case 1:
		return &matches[0], diags
	default:
		diags.AddError(
			"Multiple matching "+typeName+" resources",
			fmt.Sprintf("Found %d %s resources matching the given identifier. Names are not unique; set uuid to select one.", len(matches), typeName),
		)
		return nil, diags
	}
}

func filterByRegion[T any](items []T, region string, regionOf func(T) string) []T {
	if region == "" {
		return items
	}

	filtered := make([]T, 0, len(items))
	for _, item := range items {
		if regionOf(item) == region {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func resourceStatus(c *client.ElementoClient, billingUUID string) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	if billingUUID == "" {
		return types.StringNull(), diags
	}

	status, err := c.GetResourceStatus(billingUUID)
	if err != nil {
		diags.AddError(
			"Error Reading Resource Status",
			"Could not read billing status: "+err.Error(),
		)
		return types.StringNull(), diags
	}

	return types.StringValue(*status), diags
}
