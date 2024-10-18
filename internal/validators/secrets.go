package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// secretResourceModel maps the resource schema data.
type environmentsecretResourceModel struct {
	ProjectID         types.String          `tfsdk:"project_id"`
	Name              types.String          `tfsdk:"name"`
	Value             types.String          `tfsdk:"value"`
	Reference         types.String          `tfsdk:"reference"`
	Revision          types.Int64           `tfsdk:"revision"`
	Date              basetypes.ObjectValue `tfsdk:"date"`
	EncryptedValue    types.String          `tfsdk:"encrypted_value"`
	EncryptedRevision types.String          `tfsdk:"encrypted_revision"`
}

type certificatetsecretResourceModel struct {
	ProjectID         types.String          `tfsdk:"project_id"`
	Name              types.String          `tfsdk:"name"`
	Certificate       types.String          `tfsdk:"certificate"`
	CertificateValue  types.String          `tfsdk:"certificate_value"`
	Reference         types.String          `tfsdk:"reference"`
	Revision          types.Int64           `tfsdk:"revision"`
	Date              basetypes.ObjectValue `tfsdk:"date"`
	EncryptedValue    types.String          `tfsdk:"encrypted_value"`
	EncryptedRevision types.String          `tfsdk:"encrypted_revision"`
}

type registrySecretResourceModel struct {
	ProjectID          types.String          `tfsdk:"project_id"`
	Name               types.String          `tfsdk:"name"`
	Reference          types.String          `tfsdk:"reference"`
	RegistryUrl        types.String          `tfsdk:"registry_url"`
	AuthToken          types.String          `tfsdk:"auth_token"`
	EncryptedAuthToken types.String          `tfsdk:"encrypted_auth_token"`
	EncryptedRevision  types.String          `tfsdk:"encrypted_revision"`
	Revision           types.Int64           `tfsdk:"revision"`
	Date               basetypes.ObjectValue `tfsdk:"date"`
}

type SecretType int

const (
	Environment = iota
	Certificate
	Registry
)

type secretValidator struct {
	kind SecretType
}

func SecretValidator(kind SecretType) validator.String {
	return &secretValidator{kind}
}
func (v *secretValidator) Description(ctx context.Context) string {
	return "Validates one of plain text or encrypted values are present"
}

func (v *secretValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *secretValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	switch v.kind {
	case Environment:
		var config environmentsecretResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if resp.Diagnostics.HasError() {
			return
		}

		PlaintextValue := !config.Value.IsNull()
		hasEncryptedValue := !config.EncryptedValue.IsNull()
		hasEncryptedRevision := !config.EncryptedRevision.IsNull()

		if PlaintextValue && (hasEncryptedValue || hasEncryptedRevision) {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Both plain text value and encrypted value (or revision) cannot be set simultaneously.",
			)
		} else if !PlaintextValue && (!hasEncryptedValue || !hasEncryptedRevision) {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Either plain text value or both encrypted value and encrypted revision must be provided.",
			)
		} else if hasEncryptedValue != hasEncryptedRevision {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Both encrypted value and encrypted revision must be provided together.",
			)
		}
	case Certificate:
		var config certificatetsecretResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if resp.Diagnostics.HasError() {
			return
		}

		PlaintextValue := !config.CertificateValue.IsNull()
		hasEncryptedValue := !config.EncryptedValue.IsNull()
		hasEncryptedRevision := !config.EncryptedRevision.IsNull()

		if PlaintextValue && (hasEncryptedValue || hasEncryptedRevision) {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Both plain text value and encrypted value (or revision) cannot be set simultaneously.",
			)
		} else if !PlaintextValue && (!hasEncryptedValue || !hasEncryptedRevision) {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Either plain text value or both encrypted value and encrypted revision must be provided.",
			)
		} else if hasEncryptedValue != hasEncryptedRevision {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Both encrypted value and encrypted revision must be provided together.",
			)
		}
	case Registry:
		var config registrySecretResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if resp.Diagnostics.HasError() {
			return
		}

		PlaintextValue := !config.AuthToken.IsNull()
		hasEncryptedValue := !config.EncryptedAuthToken.IsNull()
		hasEncryptedRevision := !config.EncryptedRevision.IsNull()

		if PlaintextValue && (hasEncryptedValue || hasEncryptedRevision) {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Both plain text value and encrypted value (or revision) cannot be set simultaneously.",
			)
		} else if !PlaintextValue && (!hasEncryptedValue || !hasEncryptedRevision) {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Either plain text value or both encrypted value and encrypted revision must be provided.",
			)
		} else if hasEncryptedValue != hasEncryptedRevision {
			resp.Diagnostics.AddError(
				"Invalid Secret Configuration",
				"Both encrypted value and encrypted revision must be provided together.",
			)
		}
	default:
		resp.Diagnostics.AddError(
			"Unsupported Secret Type",
			fmt.Sprintf("The secret type %v is not supported", v.kind),
		)
	}
}
