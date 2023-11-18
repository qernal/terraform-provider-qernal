package resources

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	openapiclient "github.com/qernal/openapi-chaos-go-client"
	qernalclient "terraform-provider-qernal/internal/client"
)

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &secretResource{}
	_ resource.ResourceWithConfigure = &secretResource{}
)

// NewSecretResource is a helper function to simplify the provider implementation.
func NewSecretResource() resource.Resource {
	return &secretResource{}
}

// secretResource is the resource implementation.
type secretResource struct {
	client qernalclient.QernalAPIClient
}

func (r *secretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Metadata returns the resource type name.
func (r *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

// Schema defines the schema for the resource.
func (r *secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"payload": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"environment_value": schema.StringAttribute{
						Optional: true,
					},
					"registry": schema.StringAttribute{
						Optional: true,
					},
					"registry_value": schema.StringAttribute{
						Optional: true,
					},
					"certificate": schema.StringAttribute{
						Optional: true,
					},
					"certificate_value": schema.StringAttribute{
						Optional: true,
					},
				},
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
func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan secretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new secret
	secretType := openapiclient.SecretCreateType(plan.Type.ValueString())
	payload := openapiclient.SecretCreatePayload{}
	switch secretType {
	case openapiclient.SECRETCREATETYPE_REGISTRY:
		registry := *plan.Payload.Registry
		registryValue := *plan.Payload.RegistryValue
		payload.SecretRegistry = openapiclient.NewSecretRegistry(registry, registryValue)
	case openapiclient.SECRETCREATETYPE_CERTIFICATE:
		certificate := *plan.Payload.Certificate
		certificateValue := *plan.Payload.CertificateValue
		payload.SecretCertificate = openapiclient.NewSecretCertificate(certificate, certificateValue)
	case openapiclient.SECRETCREATETYPE_ENVIRONMENT:
		environmentValue := *plan.Payload.EnvironmentValue
		payload.SecretEnvironment = openapiclient.NewSecretEnvironment(environmentValue)
	}

	// get dek encryption key
	keyRes, _, err := r.client.SecretsAPI.ProjectsSecretsList(ctx, plan.ProjectID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not create Secret, unexpected error: "+err.Error())
		return
	}
	keys := keyRes.Data
	dekKeys := []openapiclient.SecretMetaResponse{}
	for _, d := range keys {
		if d.Type == openapiclient.SECRETMETATYPE_DEK {
			dekKeys = append(dekKeys, d)
		}
	}
	if len(dekKeys) < 1 {
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not create Secret, unexpected error: "+err.Error())
		return
	}

	encryption := fmt.Sprintf(`keys/dek/%d`, dekKeys[0].Revision)

	secret, httpRes, err := r.client.SecretsAPI.ProjectsSecretsCreate(ctx, plan.ProjectID.ValueString()).SecretBody(openapiclient.SecretBody{
		Name:       plan.Name.ValueString(),
		Type:       secretType,
		Payload:    payload,
		Encryption: encryption,
	}).Execute()
	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not create Secret, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	plan.Name = types.StringValue(secret.Name)
	plan.Type = types.StringValue(string(secret.Type))

	plan.Revision = types.Int64Value(int64(secret.Revision))

	date := resourceDate{
		CreatedAt: secret.Date.CreatedAt,
		UpdatedAt: secret.Date.UpdatedAt,
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
func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state secretResourceModel
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

	state.Type = types.StringValue(string(secret.Type))

	planPayload := payloadObj{}
	switch secret.Type {
	case openapiclient.SecretMetaType(openapiclient.SECRETCREATETYPE_REGISTRY):
		planPayload.Registry = &secret.Payload.SecretMetaResponseRegistryPayload.Registry
	case openapiclient.SecretMetaType(openapiclient.SECRETCREATETYPE_CERTIFICATE):
		planPayload.Certificate = &secret.Payload.SecretMetaResponseCertificatePayload.Certificate

	}
	state.Payload = planPayload

	state.Revision = types.Int64Value(int64(secret.Revision))

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
func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan secretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing secret
	secretType := openapiclient.SecretCreateType(plan.Type.ValueString())
	payload := openapiclient.SecretCreatePayload{}
	switch secretType {
	case openapiclient.SECRETCREATETYPE_REGISTRY:
		registry := *plan.Payload.Registry
		registryValue := *plan.Payload.RegistryValue
		payload.SecretRegistry = openapiclient.NewSecretRegistry(registry, registryValue)
	case openapiclient.SECRETCREATETYPE_CERTIFICATE:
		certificate := *plan.Payload.Certificate
		certificateValue := *plan.Payload.CertificateValue
		payload.SecretCertificate = openapiclient.NewSecretCertificate(certificate, certificateValue)
	case openapiclient.SECRETCREATETYPE_ENVIRONMENT:
		environmentValue := *plan.Payload.EnvironmentValue
		payload.SecretEnvironment = openapiclient.NewSecretEnvironment(environmentValue)
	}

	// get latest encryption key
	keyRes, _, err := r.client.SecretsAPI.ProjectsSecretsList(ctx, plan.ProjectID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not create Secret, unexpected error: "+err.Error())
		return
	}
	keys := keyRes.Data
	dekKeys := []openapiclient.SecretMetaResponse{}
	for _, d := range keys {
		if d.Type == openapiclient.SECRETMETATYPE_DEK {
			dekKeys = append(dekKeys, d)
		}
	}
	if len(dekKeys) < 1 {
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not create Secret, unexpected error: "+err.Error())
		return
	}

	encryption := fmt.Sprintf(`keys/dek/%d`, dekKeys[0].Revision)

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
	secret, _, err := r.client.SecretsAPI.ProjectsSecretsGet(ctx, plan.ProjectID.ValueString(), plan.Name.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Secret",
			"Could not read Secret Name "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.Name = types.StringValue(secret.Name)

	plan.Type = types.StringValue(string(secret.Type))

	planPayload := payloadObj{}
	switch secret.Type {
	case openapiclient.SecretMetaType(openapiclient.SECRETCREATETYPE_REGISTRY):
		planPayload.Registry = &secret.Payload.SecretMetaResponseRegistryPayload.Registry
	case openapiclient.SecretMetaType(openapiclient.SECRETCREATETYPE_CERTIFICATE):
		planPayload.Certificate = &secret.Payload.SecretMetaResponseCertificatePayload.Certificate

	}

	plan.Payload = planPayload

	plan.Revision = types.Int64Value(int64(secret.Revision))

	date := resourceDate{
		CreatedAt: secret.Date.CreatedAt,
		UpdatedAt: secret.Date.UpdatedAt,
	}
	plan.Date = date.GetDateObject()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state secretResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing secret
	_, _, err := r.client.SecretsAPI.ProjectsSecretsDelete(ctx, state.ProjectID.ValueString(), state.Name.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting secret",
			"Could not delete secret, unexpected error: "+err.Error(),
		)
		return
	}
}

// secretResourceModel maps the resource schema data.
type secretResourceModel struct {
	ProjectID types.String          `tfsdk:"project_id"`
	Name      types.String          `tfsdk:"name"`
	Type      types.String          `tfsdk:"type"`
	Payload   payloadObj            `tfsdk:"payload"`
	Revision  types.Int64           `tfsdk:"revision"`
	Date      basetypes.ObjectValue `tfsdk:"date"`
}

type payloadObj struct {
	EnvironmentValue *string `tfsdk:"environment_value"`

	Certificate      *string `tfsdk:"certificate"`
	CertificateValue *string `tfsdk:"certificate_value"`

	Registry      *string `tfsdk:"registry"`
	RegistryValue *string `tfsdk:"registry_value"`
}
