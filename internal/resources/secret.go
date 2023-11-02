package resources

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	openapiclient "github.com/qernal/openapi-chaos-go-client"
	qernalclient "qernal-terraform-provider/internal/client"
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
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"payload": schema.SingleNestedAttribute{
				Required: true,
			},
			"encryption": schema.StringAttribute{
				Optional: false,
			},
			"revision": schema.StringAttribute{
				Computed: true,
			},
			"date": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"created_at": schema.StringAttribute{
						Computed: true,
					},
					"updated_at": schema.StringAttribute{
						Computed: true,
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
		registry := plan.Payload.Attributes()["registry"].String()
		registryValue := plan.Payload.Attributes()["registry_value"].String()
		payload.SecretRegistry = openapiclient.NewSecretRegistry(registry, registryValue)
	case openapiclient.SECRETCREATETYPE_CERTIFICATE:
		certificate := plan.Payload.Attributes()["certificate"].String()
		certificateValue := plan.Payload.Attributes()["certificate_value"].String()
		payload.SecretCertificate = openapiclient.NewSecretCertificate(certificate, certificateValue)
	case openapiclient.SECRETCREATETYPE_ENVIRONMENT:
		environmentValue := plan.Payload.Attributes()["environment_value"].String()
		payload.SecretEnvironment = openapiclient.NewSecretEnvironment(environmentValue)
	}
	secret, _, err := r.client.SecretsAPI.ProjectsSecretsCreate(ctx, plan.ProjectID.ValueString()).SecretBody(openapiclient.SecretBody{
		Name:       plan.Name.ValueString(),
		Type:       secretType,
		Payload:    payload,
		Encryption: plan.Encryption.ValueString(),
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Project",
			"Could not create Project, unexpected error: "+err.Error())
		return
	}

	plan.Name = types.StringValue(secret.Name)
	plan.Type = types.StringValue(string(secret.Type))

	planPayload := basetypes.ObjectValue{}
	switch secretType {
	case openapiclient.SECRETCREATETYPE_REGISTRY:
		planPayload, _ = types.ObjectValue(
			map[string]attr.Type{
				"registry":       types.StringType,
				"registry_value": types.StringType,
			},
			map[string]attr.Value{
				"registry":       types.StringValue(plan.Payload.Attributes()["registry"].String()),
				"registry_value": types.StringValue(plan.Payload.Attributes()["registry_value"].String()),
			},
		)
	case openapiclient.SECRETCREATETYPE_CERTIFICATE:
		planPayload, _ = types.ObjectValue(
			map[string]attr.Type{
				"certificate":       types.StringType,
				"certificate_value": types.StringType,
			},
			map[string]attr.Value{
				"certificate":       types.StringValue(plan.Payload.Attributes()["certificate"].String()),
				"certificate_value": types.StringValue(plan.Payload.Attributes()["certificate_value"].String()),
			},
		)

	}
	plan.Payload = planPayload

	plan.Revision = types.Int64Value(int64(secret.Revision))

	date, _ := types.ObjectValue(
		map[string]attr.Type{
			"created_at": types.StringType,
			"updated_at": types.StringType,
		},
		map[string]attr.Value{
			"created_at": types.StringValue(secret.Date.CreatedAt),
			"updated_at": types.StringValue(secret.Date.UpdatedAt),
		},
	)
	plan.Date = date

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

	planPayload := basetypes.ObjectValue{}
	switch secret.Type {
	case openapiclient.SecretMetaType((openapiclient.SECRETCREATETYPE_REGISTRY)):
		planPayload, _ = types.ObjectValue(
			map[string]attr.Type{
				"registry": types.StringType,
			},
			map[string]attr.Value{
				"registry": types.StringValue(secret.Payload.SecretMetaResponseRegistryPayload.Registry),
			},
		)
	case openapiclient.SecretMetaType(openapiclient.SECRETCREATETYPE_CERTIFICATE):
		planPayload, _ = types.ObjectValue(
			map[string]attr.Type{
				"certificate": types.StringType,
			},
			map[string]attr.Value{
				"certificate": types.StringValue(secret.Payload.SecretMetaResponseCertificatePayload.Certificate),
			},
		)

	}
	state.Payload = planPayload

	state.Revision = types.Int64Value(int64(secret.Revision))

	state.Date, _ = types.ObjectValue(
		map[string]attr.Type{
			"created_at": types.StringType,
			"updated_at": types.StringType,
		},
		map[string]attr.Value{
			"created_at": types.StringValue(secret.Date.CreatedAt),
			"updated_at": types.StringValue(secret.Date.UpdatedAt),
		},
	)

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
		registry := plan.Payload.Attributes()["registry"].String()
		registryValue := plan.Payload.Attributes()["registry_value"].String()
		payload.SecretRegistry = openapiclient.NewSecretRegistry(registry, registryValue)
	case openapiclient.SECRETCREATETYPE_CERTIFICATE:
		certificate := plan.Payload.Attributes()["certificate"].String()
		certificateValue := plan.Payload.Attributes()["certificate_value"].String()
		payload.SecretCertificate = openapiclient.NewSecretCertificate(certificate, certificateValue)
	case openapiclient.SECRETCREATETYPE_ENVIRONMENT:
		environmentValue := plan.Payload.Attributes()["environment_value"].String()
		payload.SecretEnvironment = openapiclient.NewSecretEnvironment(environmentValue)
	}
	_, _, err := r.client.SecretsAPI.ProjectsSecretsUpdate(ctx, plan.ProjectID.ValueString(), plan.Name.ValueString()).SecretBodyPatch(openapiclient.SecretBodyPatch{
		Type:    secretType,
		Payload: payload,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating secret",
			"Could not update secret, unexpected error: "+err.Error(),
		)
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

	planPayload := basetypes.ObjectValue{}
	switch secret.Type {
	case openapiclient.SecretMetaType((openapiclient.SECRETCREATETYPE_REGISTRY)):
		planPayload, _ = types.ObjectValue(
			map[string]attr.Type{
				"registry": types.StringType,
			},
			map[string]attr.Value{
				"registry": types.StringValue(secret.Payload.SecretMetaResponseRegistryPayload.Registry),
			},
		)
	case openapiclient.SecretMetaType(openapiclient.SECRETCREATETYPE_CERTIFICATE):
		planPayload, _ = types.ObjectValue(
			map[string]attr.Type{
				"certificate": types.StringType,
			},
			map[string]attr.Value{
				"certificate": types.StringValue(secret.Payload.SecretMetaResponseCertificatePayload.Certificate),
			},
		)

	}
	plan.Payload = planPayload

	plan.Revision = types.Int64Value(int64(secret.Revision))

	plan.Date, _ = types.ObjectValue(
		map[string]attr.Type{
			"created_at": types.StringType,
			"updated_at": types.StringType,
		},
		map[string]attr.Value{
			"created_at": types.StringValue(secret.Date.CreatedAt),
			"updated_at": types.StringValue(secret.Date.UpdatedAt),
		},
	)

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
	ProjectID  types.String          `tfsdk:"project_id"`
	Name       types.String          `tfsdk:"name"`
	Type       types.String          `tfsdk:"type"`
	Payload    basetypes.ObjectValue `tfsdk:"payload"`
	Encryption types.String          `tfsdk:"encryption"`
	Revision   types.Int64           `tfsdk:"revision"`
	Date       basetypes.ObjectValue `tfsdk:"date"`
}
