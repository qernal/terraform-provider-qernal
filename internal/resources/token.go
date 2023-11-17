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
	_ resource.Resource              = &tokenResource{}
	_ resource.ResourceWithConfigure = &tokenResource{}
)

// NewtokenResource is a helper function to simplify the provider implementation.
func NewTokenResource() resource.Resource {
	return &tokenResource{}
}

// tokenResource is the resource implementation.
type tokenResource struct {
	client qernalclient.QernalAPIClient
}

func (r *tokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *tokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

// Schema defines the schema for the resource.
func (r *tokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"user_id": schema.StringAttribute{
				Computed: true,
				Required: false,
			},
			"expiry_at": schema.StringAttribute{
				Computed: true,
				Required: false,
			},
			"expiry_duration": schema.Int64Attribute{
				Required: true,
			},
			"token": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"date": schema.SingleNestedAttribute{
				Computed: true,
				Required: false,
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
func (r *tokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan tokenResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new token
	token, _, err := r.client.TokensAPI.AuthTokensCreate(ctx).AuthTokenBody(openapiclient.AuthTokenBody{
		Name:           plan.Name.ValueString(),
		ExpiryDuration: int32(plan.ExpiryDuration.ValueInt64()),
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating token",
			"Could not create token, unexpected error: "+err.Error())
		return
	}

	plan.ID = types.StringValue(token.Id)
	plan.Name = types.StringValue(token.Name)
	plan.UserID = types.StringValue(token.UserId)
	plan.ExpiryAt = types.StringValue(*token.ExpiryAt)
	plan.Token = types.StringValue(*token.Token)

	date := resourceDate{
		CreatedAt: token.Date.CreatedAt,
		UpdatedAt: token.Date.UpdatedAt,
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
func (r *tokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state tokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed token value from qernal
	token, _, err := r.client.TokensAPI.AuthTokensGet(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading token",
			"Could not read token Name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(token.Name)
	state.UserID = types.StringValue(token.UserId)
	state.ExpiryAt = types.StringValue(*token.ExpiryAt)

	date := resourceDate{
		CreatedAt: token.Date.CreatedAt,
		UpdatedAt: token.Date.UpdatedAt,
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
func (r *tokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan tokenResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing token

	expiryDuration := int32(plan.ExpiryDuration.ValueInt64())
	_, _, err := r.client.TokensAPI.AuthTokensUpdate(ctx, plan.ID.ValueString()).AuthTokenPatch(openapiclient.AuthTokenPatch{
		Name:           plan.Name.ValueStringPointer(),
		ExpiryDuration: &expiryDuration,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating token",
			"Could not update token, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch updated Project
	token, _, err := r.client.TokensAPI.AuthTokensGet(ctx, plan.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading token",
			"Could not read token "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.Name = types.StringValue(token.Name)
	plan.ExpiryAt = types.StringValue(*token.ExpiryAt)

	date := resourceDate{
		CreatedAt: token.Date.CreatedAt,
		UpdatedAt: token.Date.UpdatedAt,
	}
	plan.Date = date.GetDateObject()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *tokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state tokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing token
	_, _, err := r.client.TokensAPI.AuthTokensDelete(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting token",
			"Could not delete token, unexpected error: "+err.Error(),
		)
		return
	}
}

// tokenResourceModel maps the resource schema data.
type tokenResourceModel struct {
	ID             types.String          `tfsdk:"id"`
	Name           types.String          `tfsdk:"name"`
	UserID         types.String          `tfsdk:"user_id"`
	ExpiryAt       types.String          `tfsdk:"expiry_at"`
	ExpiryDuration types.Int64           `tfsdk:"expiry_duration"`
	Token          types.String          `tfsdk:"token"`
	Date           basetypes.ObjectValue `tfsdk:"date"`
}
