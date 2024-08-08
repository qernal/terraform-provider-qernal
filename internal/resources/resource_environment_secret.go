package resources

import (
	"context"
	"fmt"
	"terraform-provider-qernal/internal/client"
	qernalclient "terraform-provider-qernal/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	openapiclient "github.com/qernal/openapi-chaos-go-client"
)

var (
	_ resource.Resource              = &environmentsecretResource{}
	_ resource.ResourceWithConfigure = &environmentsecretResource{}
)

func NewenvironmentsecretResource() resource.Resource {
	return &environmentsecretResource{}

}

type environmentsecretResource struct {
	client qernalclient.QernalAPIClient
}

func (r *environmentsecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *environmentsecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_environment"
}

// Schema defines the schema for the resource.
func (r *environmentsecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the envrionment variable. e.g( PORT)",
			},

			"value": schema.StringAttribute{
				Required:    true,
				Description: "Value of the environment variable",
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

// Create creates the resource and sets the initial Terraform state.
func (r *environmentsecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan environmentsecretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch dek key
	keyRes, err := r.client.FetchDek(ctx, plan.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not obtain encryption key: "+err.Error())
		return
	}
	// Create new secret
	secretType := openapiclient.SecretCreateType(openapiclient.SECRETCREATETYPE_ENVIRONMENT)
	payload := openapiclient.SecretCreatePayload{}

	// encrypt secret
	secret := plan.Value.ValueString()

	encryptedSecret, err := client.EncryptLocalSecret(keyRes.Payload.SecretMetaResponseDek.Public, secret)
	if err != nil {
		resp.Diagnostics.AddError("unable to encrypt local secret", "encryption failed with:"+err.Error())
	}

	payload.SecretEnvironment = openapiclient.NewSecretEnvironment(encryptedSecret)

	encryptionRef := fmt.Sprintf(`keys/dek/%d`, keyRes.Revision)

	secretBody := openapiclient.NewSecretBody(plan.Name.ValueString(), secretType, payload, encryptionRef)

	secretRes, httpRes, err := r.client.SecretsAPI.ProjectsSecretsCreate(ctx, plan.ProjectID.ValueString()).SecretBody(*secretBody).Execute()

	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not create Secret, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	plan.Name = types.StringValue(secretRes.Name)
	plan.Reference = types.StringValue(fmt.Sprintf("projects:%s/%s@%d", plan.ProjectID, plan.Name, plan.Revision))
	plan.Revision = types.Int64Value(int64(secretRes.Revision))

	date := resourceDate{
		CreatedAt: secretRes.Date.CreatedAt,
		UpdatedAt: secretRes.Date.UpdatedAt,
	}
	plan.Date = date.GetDateObject()

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read refreshes the Terraform state with the latest data.
func (r *environmentsecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state environmentsecretResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed secret value from qernal
	secret, _, err := r.client.SecretsAPI.ProjectsSecretsGet(ctx, state.ProjectID.ValueString(), state.Name.ValueString()).Execute()
	if err != nil {

		resp.Diagnostics.AddError(
			"Error Reading Secret",
			"Could not read Secret Name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(secret.Name)

	date := resourceDate{
		CreatedAt: secret.Date.CreatedAt,
		UpdatedAt: secret.Date.UpdatedAt,
	}
	state.Date = date.GetDateObject()

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *environmentsecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan environmentsecretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing secret
	secretType := openapiclient.SecretCreateType(openapiclient.SECRETCREATETYPE_ENVIRONMENT)
	payload := openapiclient.SecretCreatePayload{}

	secret := plan.Value.ValueString()

	// Fetch dek key
	keyRes, err := r.client.FetchDek(ctx, plan.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not obtain encryption key: "+err.Error())
		return
	}

	// encrypt secret

	encryptedSecret, err := client.EncryptLocalSecret(keyRes.Payload.SecretMetaResponseDek.Public, secret)
	if err != nil {
		resp.Diagnostics.AddError("unable to encrypt local secret", "encryption failed with:"+err.Error())
	}

	payload.SecretEnvironment = openapiclient.NewSecretEnvironment(encryptedSecret)

	encryption := fmt.Sprintf(`keys/dek/%d`, keyRes.Revision)

	_, httpRes, err := r.client.SecretsAPI.ProjectsSecretsUpdate(ctx, plan.ProjectID.ValueString(), plan.Name.ValueString()).SecretBodyPatch(openapiclient.SecretBodyPatch{
		Type:       secretType,
		Payload:    payload,
		Encryption: encryption,
	}).Execute()
	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error updating Secret",
			"Could not update Secret, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	// Fetch updated Project
	secretRes, _, err := r.client.SecretsAPI.ProjectsSecretsGet(ctx, plan.ProjectID.ValueString(), plan.Name.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Secret",
			"Could not read Secret Name "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.Name = types.StringValue(secretRes.Name)

	// plan.Payload = planPayload

	plan.Revision = types.Int64Value(int64(secretRes.Revision))

	date := resourceDate{
		CreatedAt: secretRes.Date.CreatedAt,
		UpdatedAt: secretRes.Date.UpdatedAt,
	}
	plan.Date = date.GetDateObject()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *environmentsecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state environmentsecretResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing secret
	_, httpRes, err := r.client.SecretsAPI.ProjectsSecretsDelete(ctx, state.ProjectID.ValueString(), state.Name.ValueString()).Execute()
	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)

		ctx = tflog.SetField(ctx, "raw http response", resData)
		tflog.Error(ctx, " deletion failed")
		resp.Diagnostics.AddError(
			"Error Deleting secret",
			"Could not delete secret, unexpected error: "+err.Error(),
		)
		return
	}
}

// secretResourceModel maps the resource schema data.
type environmentsecretResourceModel struct {
	ProjectID types.String          `tfsdk:"project_id"`
	Name      types.String          `tfsdk:"name"`
	Value     types.String          `tfsdk:"value"`
	Reference types.String          `tfsdk:"reference"`
	Revision  types.Int64           `tfsdk:"revision"`
	Date      basetypes.ObjectValue `tfsdk:"date"`
}
