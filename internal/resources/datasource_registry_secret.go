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
	_ datasource.DataSource = &registryDataSource{}
)

func NewregistryDataSource() datasource.DataSource {
	return &registryDataSource{}
}

type registryDataSource struct {
	client qernalclient.QernalAPIClient
}

func (r *registryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *registryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_registry"
}

// Schema defines the schema for the data source.
func (d *registryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
			},

			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the registry secret",
			},

			"registry": schema.StringAttribute{
				Computed:    true,
				Description: "url of the registry",
			},

			"revision": schema.Int64Attribute{
				Computed: true,
				Required: false,
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
func (d *registryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data registrysecretDataSourceModel

	// Read Terraform configuration data into the model

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	secret, httpRes, err := d.client.SecretsAPI.ProjectsSecretsGet(ctx, data.ProjectID.ValueString(), data.Name.ValueString()).Execute()

	if err != nil {

		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error retreivng Secret",
			"Could not get Secret, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	data.Name = types.StringValue(secret.Name)

	data.ProjectID = types.StringValue(data.ProjectID.ValueString())

	data.Registry = types.StringValue(secret.Payload.SecretMetaResponseRegistryPayload.Registry)

	data.Revision = types.Int64Value(int64(secret.Revision))

	date := resourceDate{
		CreatedAt: secret.Date.CreatedAt,
		UpdatedAt: secret.Date.UpdatedAt,
	}
	data.Date = date.GetDateObject()

	// Set refreshed data
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

type registrysecretDataSourceModel struct {
	ProjectID types.String          `tfsdk:"project_id"`
	Name      types.String          `tfsdk:"name"`
	Registry  types.String          `tfsdk:"registry"`
	Revision  types.Int64           `tfsdk:"revision"`
	Date      basetypes.ObjectValue `tfsdk:"date"`
}
