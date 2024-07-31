package resources

import (
	"context"
	"fmt"
	"strings"
	qernalclient "terraform-provider-qernal/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	openapi_chaos_client "github.com/qernal/openapi-chaos-go-client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &providerDataSource{}
)

func NewproviderDataSource() datasource.DataSource {
	return &providerDataSource{}
}

type providerDataSource struct {
	client qernalclient.QernalAPIClient
}

func (r *providerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(qernalclient.QernalAPIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.QernalAPIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

// Metadata returns the data source type name.
func (d *providerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

// Schema defines the schema for the resource.
func (r *providerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of provider",
				Required:    false,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of provider",
				Required:    true,
			},
			"continents": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Available continents",
			},
			"countries": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Available countries",
			},
			"cities": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Available cities",
			},
		},
	}
}

// Read refreshes the Terraform data with the latest data.
func (d *providerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data providerDataSourceModel

	// Read Terraform configuration data into the model

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	providers, httpRes, err := d.client.ProvidersAPI.ProvidersList(ctx).Execute()

	if err != nil {

		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error Fetching providers",
			"Could not retreive providers, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	var qernalProvider openapi_chaos_client.Provider
	for _, provider := range providers.Data {
		if strings.EqualFold(data.Name.ValueString(), provider.Name) {
			qernalProvider = provider
			break
		}
	}

	if qernalProvider.Name == "" {
		// provider not found
		resp.Diagnostics.AddError("Unable to find provider", fmt.Sprintf("Unable to locate provider %s, this provider may not exist or has been removed", data.Name))
		return
	}

	// Update resource state with updated items and timestamp

	qernalContinents := make([]types.String, 0, len(qernalProvider.Locations.Continents))
	for _, continent := range qernalProvider.Locations.Continents {
		qernalContinents = append(qernalContinents, types.StringValue(continent))
	}

	qernalCountries := make([]types.String, 0, len(qernalProvider.Locations.Countries))

	for _, country := range qernalProvider.Locations.Countries {
		qernalCountries = append(qernalContinents, types.StringValue(country))
	}

	qernalCities := make([]types.String, 0, len(qernalProvider.Locations.Cities))

	for _, city := range qernalProvider.Locations.Cities {
		qernalCities = append(qernalCities, types.StringValue(city))
	}

	data.Name = types.StringValue(qernalProvider.Name)
	data.ID = types.StringValue(qernalProvider.Id)
	data.Continents = qernalContinents
	data.Countries = qernalCountries
	data.Cities = qernalCities

	// Set refreshed data
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

type providerDataSourceModel struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	Continents []types.String `tfsdk:"continents"`
	Countries  []types.String `tfsdk:"countries"`
	Cities     []types.String `tfsdk:"cities"`
}
