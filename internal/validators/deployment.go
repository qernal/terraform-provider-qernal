package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FunctionResourceModel maps the resource schema data.
type FunctionResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	Revision     types.String   `tfsdk:"revision"`
	ProjectID    types.String   `tfsdk:"project_id"`
	Name         types.String   `tfsdk:"name"`
	Version      types.String   `tfsdk:"version"`
	Description  types.String   `tfsdk:"description"`
	Image        types.String   `tfsdk:"image"`
	FunctionType types.String   `tfsdk:"type"`
	Size         Size           `tfsdk:"size"`
	Port         types.Int64    `tfsdk:"port"`
	Route        []Route        `tfsdk:"route"`
	Scaling      Scaling        `tfsdk:"scaling"`
	Deployment   []Deployment   `tfsdk:"deployment"`
	Secret       []Secret       `tfsdk:"secret"`
	Compliance   []types.String `tfsdk:"compliance"`
}

// Size represents the size configuration for the function
type Size struct {
	CPU    types.Int64 `tfsdk:"cpu"`
	Memory types.Int64 `tfsdk:"memory"`
}

// Route represents a route configuration for the function
type Route struct {
	Path    types.String   `tfsdk:"path"`
	Methods []types.String `tfsdk:"methods"`
	Weight  types.Int64    `tfsdk:"weight"`
}

// Scaling represents the scaling configuration for the function
type Scaling struct {
	Type types.String `tfsdk:"type"`
	Low  types.Int64  `tfsdk:"low"`
	High types.Int64  `tfsdk:"high"`
}

// Deployment represents a deployment configuration for the function
type Deployment struct {
	Location Location `tfsdk:"location"`
	Replicas Replicas `tfsdk:"replicas"`
}

// Location represents the location details for a deployment
type Location struct {
	ProviderID types.String `tfsdk:"provider_id"`
	Continent  types.String `tfsdk:"continent"`
	Country    types.String `tfsdk:"country"`
	City       types.String `tfsdk:"city"`
}

// Replicas represents the replicas configuration for a deployment
type Replicas struct {
	Min      types.Int64 `tfsdk:"min"`
	Max      types.Int64 `tfsdk:"max"`
	Affinity Affinity    `tfsdk:"affinity"`
}

// Affinity represents the affinity configuration for replicas
type Affinity struct {
	Cluster types.Bool `tfsdk:"cluster"`
	Cloud   types.Bool `tfsdk:"cloud"`
}

// Secret represents a secret configuration for the function
type Secret struct {
	Name      types.String `tfsdk:"name"`
	Reference types.String `tfsdk:"reference"`
}

type locationValidator struct{}

func LocationValidator() validator.Object {
	return &locationValidator{}
}

func (v *locationValidator) Description(ctx context.Context) string {
	return "validates that provider_id is present and at least one of continent, country, or city is specified"
}

func (v *locationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *locationValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	var config FunctionResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Deployment[0].Location.ProviderID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("provider_id"),
			"Missing Provider ID",
			"provider_id is required",
		)
	}

	if config.Deployment[0].Location.Continent.IsNull() && config.Deployment[0].Location.Country.IsNull() && config.Deployment[0].Location.City.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Location Information",
			"At least one of continent, country, or city must be specified",
		)
	}
}
