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

func NewprojectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client qernalclient.QernalAPIClient
}

func (r *projectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the data source.
func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
			},

			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the project",
			},

			"org_id": schema.StringAttribute{
				Computed:    true,
				Description: "Organisation associated with the project",
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
func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data projectsecretDataSourceModel

	// Read Terraform configuration data into the model

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	project, httpRes, err := d.client.ProjectsAPI.ProjectsGet(ctx, data.ProjectID.ValueString()).Execute()

	if err != nil {

		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error retreiving  project",
			"Could not retrieve project, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	data.Name = types.StringValue(project.Name)

	data.ProjectID = types.StringValue(data.ProjectID.ValueString())

	data.OrgID = types.StringValue(project.Id)

	date := resourceDate{
		CreatedAt: project.Date.CreatedAt,
		UpdatedAt: project.Date.UpdatedAt,
	}
	data.Date = date.GetDateObject()

	// Set refreshed data
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

type projectsecretDataSourceModel struct {
	ProjectID types.String          `tfsdk:"project_id"`
	Name      types.String          `tfsdk:"name"`
	OrgID     types.String          `tfsdk:"org_id"`
	Date      basetypes.ObjectValue `tfsdk:"date"`
}
