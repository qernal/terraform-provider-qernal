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
	_ datasource.DataSource = &certificateDataSource{}
)

func NewcertificateDataSource() datasource.DataSource {
	return &certificateDataSource{}
}

type certificateDataSource struct {
	client qernalclient.QernalAPIClient
}

func (r *certificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *certificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_certificate"
}

// Schema defines the schema for the data source.
func (d *certificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
			},

			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the certficate",
			},

			"certificate": schema.StringAttribute{
				Computed:    true,
				Description: "Public key of the certificate",
			},
			"reference": schema.StringAttribute{
				Computed:    true,
				Description: "reference attribute of the secret",
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
func (d *certificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data certificatesecretDataSourceModel

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
			"Error creating Secret",
			"Could not create Secret, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	data.Name = types.StringValue(secret.Name)
	data.ProjectID = types.StringValue(data.ProjectID.ValueString())
	data.Reference = types.StringValue(fmt.Sprintf("projects:%s/%s@%d", data.ProjectID, data.Name, data.Revision))
	data.Revision = types.Int64Value(int64(secret.Revision))
	data.Certificate = types.StringValue(secret.Payload.SecretMetaResponseCertificatePayload.Certificate)
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

type certificatesecretDataSourceModel struct {
	ProjectID   types.String          `tfsdk:"project_id"`
	Name        types.String          `tfsdk:"name"`
	Certificate types.String          `tfsdk:"certificate"`
	Reference   types.String          `tfsdk:"reference"`
	Revision    types.Int64           `tfsdk:"revision"`
	Date        basetypes.ObjectValue `tfsdk:"date"`
}
