package resources

import (
	"context"
	"fmt"
	qernalclient "terraform-provider-qernal/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &projectDataSource{}
)

func NeworganisationDataSource() datasource.DataSource {
	return &organisationDataSource{}
}

type organisationDataSource struct {
	client qernalclient.QernalAPIClient
}

func (r *organisationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *organisationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organisation"
}

// Schema defines the schema for the data source.
func (d *organisationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the organisation",
			},

			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the organisation",
			},

			"user_id": schema.StringAttribute{
				Computed:    true,
				Description: "User ID associated with the organisation",
			},
			"date": schema.SingleNestedAttribute{
				Computed: true,
				Required: false,
				Attributes: map[string]schema.Attribute{
					"created_at": schema.StringAttribute{
						Optional: true,
					},
					"updated_at": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform data with the latest data.
func (d *organisationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data organisationsecretDataSourceModel

	// Read Terraform configuration data into the model

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	organisations, httpRes, err := d.client.OrganisationsAPI.OrganisationsList(ctx).FName(data.Name.ValueString()).Execute()

	if err != nil {

		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error retreiving  organisation",
			"Could not retrieve organisation, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	if len(organisations.Data) <= 0 {
		resp.Diagnostics.AddError("Unable to find organisation", "Unable to find organisation with name "+data.OrganisationID.String())
		return
	}

	organisation := organisations.Data[0]

	data.Name = types.StringValue(organisation.Name)

	data.OrganisationID = types.StringValue(organisation.Id)

	data.UserID = types.StringValue(organisation.UserId)

	date := resourceDate{
		CreatedAt: organisation.Date.CreatedAt,
		UpdatedAt: organisation.Date.UpdatedAt,
	}
	data.Date = date.GetDateObject()

	// Set refreshed data
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

type organisationsecretDataSourceModel struct {
	OrganisationID types.String          `tfsdk:"id"`
	Name           types.String          `tfsdk:"name"`
	UserID         types.String          `tfsdk:"user_id"`
	Date           basetypes.ObjectValue `tfsdk:"date"`
}
