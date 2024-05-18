package resources

import (
	"context"
	"fmt"
	qernalclient "terraform-provider-qernal/internal/client"
	pkgtypes "terraform-provider-qernal/pkg/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &hostDataSource{}
)

func NewhostDataSource() datasource.DataSource {
	return &hostDataSource{}
}

type hostDataSource struct {
	client qernalclient.QernalAPIClient
}

func (r *hostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *hostDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

// Schema defines the schema for the resource.
func (r *hostDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: false,
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"hostname": schema.StringAttribute{
				Required: true,
			},
			"certificate": schema.StringAttribute{
				Required: false,
				Computed: true,
			},
			"disabled": schema.BoolAttribute{
				Required: false,
				Computed: true,
			},
			"read_only": schema.BoolAttribute{
				Required: false,
				Computed: true,
			},
			"txt_verification": schema.StringAttribute{
				Required: false,
				Computed: true,
			},
			"verified_at": schema.StringAttribute{
				Required: false,
				Computed: true,
			},
			"verification_status": schema.StringAttribute{
				Required: false,
				Computed: true,
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
func (d *hostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data hostDataSourceModel

	// Read Terraform configuration data into the model

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	host, httpRes, err := d.client.HostsAPI.ProjectsHostsGet(ctx, data.ProjectID.ValueString(), data.Name.ValueString()).Execute()

	if err != nil {

		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error Fetching Host",
			"Could not retreive Host, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	// Update resource state with updated items and timestamp
	data.Name = types.StringValue(host.Host)
	data.ID = types.StringValue(host.Id)
	data.ProjectID = types.StringValue(host.ProjectId)
	data.Certificate = types.StringValue(pkgtypes.StringValueFromPointer(host.Certificate))
	data.ReadOnly = types.BoolValue(host.ReadOnly)
	data.Disabled = types.BoolValue(host.Disabled)
	data.TxtVerification = types.StringValue(host.TxtVerification)
	data.VerifiedAt = types.StringValue(pkgtypes.StringValueFromPointer(host.VerifiedAt))
	data.VerificationStatus = types.StringValue(string(host.VerificationStatus))

	date := resourceDate{
		CreatedAt: host.Date.CreatedAt,
		UpdatedAt: host.Date.UpdatedAt,
	}
	data.Date = date.GetDateObject()

	// Set refreshed data
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

type hostDataSourceModel struct {
	ID                 types.String          `tfsdk:"id"`
	ProjectID          types.String          `tfsdk:"project_id"`
	Name               types.String          `tfsdk:"hostname "`
	Certificate        types.String          `tfsdk:"certificate"`
	ReadOnly           types.Bool            `tfsdk:"read_only"`
	Disabled           types.Bool            `tfsdk:"disabled"`
	TxtVerification    types.String          `tfsdk:"txt_verification"`
	VerifiedAt         types.String          `tfsdk:"verified_at"`
	VerificationStatus types.String          `tfsdk:"verification_status"`
	Date               basetypes.ObjectValue `tfsdk:"date"`
}
