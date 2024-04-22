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
	_ resource.Resource              = &registrySecretResource{}
	_ resource.ResourceWithConfigure = &registrySecretResource{}
)

func NewregistrySecretResource() resource.Resource {
	return &registrySecretResource{}

}

type registrySecretResource struct {
	client qernalclient.QernalAPIClient
}

func (r *registrySecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *registrySecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_secret"
}

// Schema defines the schema for the resource.
func (r *registrySecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"registry_url": schema.StringAttribute{
				Required:    true,
				Description: "The URL of the container registry (e.g., ghcr.io)",
			},

			"auth_token": schema.StringAttribute{
				Required:    true,
				Description: " authentication token for the registry",
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
func (r *registrySecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan registrySecretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch dek key
	keyRes, httpres, err := r.client.SecretsAPI.ProjectsSecretsGet(ctx, plan.ProjectID.ValueString(), "dek").Execute()
	if err != nil {
		resData, httperr := qernalclient.ParseResponseData(httpres)
		ctx = tflog.SetField(ctx, "http response", httperr)
		tflog.Error(ctx, "response from server")
		if httperr != nil {
			resp.Diagnostics.AddError(
				"Error creating Secret",
				"Could not obtain dek key, unexpected http error: "+err.Error())
			return
		}
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not obtain encryption key, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	// Create new secret
	secretType := openapiclient.SecretCreateType(openapiclient.SECRETCREATETYPE_REGISTRY)
	payload := openapiclient.SecretCreatePayload{}

	registry := plan.RegistryUrl.ValueString()

	// encrypt secret
	registryValue := plan.AuthToken.ValueString()

	encryptedToken, err := client.EncryptLocalSecret(keyRes.Payload.SecretMetaResponseDek.Public, registryValue)
	if err != nil {
		resp.Diagnostics.AddError("unable to encrypt local secret", "encryption failed with:"+err.Error())
	}

	payload.SecretRegistry = &openapiclient.SecretRegistry{
		Registry:      registry,
		RegistryValue: encryptedToken,
	}

	encryptionRef := fmt.Sprintf(`keys/dek/%d`, keyRes.Revision)

	secret, httpRes, err := r.client.SecretsAPI.ProjectsSecretsCreate(ctx, plan.ProjectID.ValueString()).SecretBody(openapiclient.SecretBody{
		Name:       plan.Name.ValueString(),
		Type:       secretType,
		Payload:    payload,
		Encryption: encryptionRef,
	}).Execute()

	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not create Secret, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	plan.Name = types.StringValue(secret.Name)

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
func (r *registrySecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state registrySecretResourceModel
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

	// state.Type = types.StringValue(string(openapiclient.SECRETCREATETYPE_REGISTRY))

	planPayload := payloadObj{}

	planPayload.Registry = &secret.Payload.SecretMetaResponseRegistryPayload.Registry

	// state.Payload = planPayload

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
func (r *registrySecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan registrySecretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing secret
	secretType := openapiclient.SecretCreateType(openapiclient.SECRETCREATETYPE_REGISTRY)
	payload := openapiclient.SecretCreatePayload{}

	registry := plan.RegistryUrl.String()

	registryValue := plan.AuthToken.String()

	payload.SecretRegistry = openapiclient.NewSecretRegistry(registry, registryValue)

	// Fetch dek key
	keyRes, httpres, err := r.client.SecretsAPI.ProjectsSecretsGet(ctx, plan.ProjectID.ValueString(), "dek").Execute()
	if err != nil {
		resData, httperr := qernalclient.ParseResponseData(httpres)
		ctx = tflog.SetField(ctx, "http response", httperr)
		tflog.Error(ctx, "response from server")
		if httperr != nil {
			resp.Diagnostics.AddError(
				"Error creating Secret",
				"Could not obtain dek key, unexpected http error: "+err.Error())
			return
		}
		resp.Diagnostics.AddError(
			"Error creating Secret",
			"Could not obtain encryption key, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

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

	// plan.Type = types.StringValue(string(secret.Type))

	planPayload := payloadObj{}

	planPayload.Registry = &secret.Payload.SecretMetaResponseRegistryPayload.Registry

	// plan.Payload = planPayload

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
func (r *registrySecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state registrySecretResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing secret
	_, _, err := r.client.SecretsAPI.ProjectsSecretsDelete(ctx, state.ProjectID.String(), state.Name.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting secret",
			"Could not delete secret, unexpected error: "+err.Error(),
		)
		return
	}
}

// secretResourceModel maps the resource schema data.
type registrySecretResourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	// Type        types.String          `tfsdk:"type"`
	// Payload     payloadObj            `tfsdk:"payload"`
	RegistryUrl types.String          `tfsdk:"registry_url"`
	AuthToken   types.String          `tfsdk:"auth_token"`
	Revision    types.Int64           `tfsdk:"revision"`
	Date        basetypes.ObjectValue `tfsdk:"date"`
}

// type payloadObj struct {
// 	Registry      *string `tfsdk:"registry"`
// 	RegistryValue *string `tfsdk:"registry_value"`
// }
