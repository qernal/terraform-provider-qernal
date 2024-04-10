package resources

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource              = &organisationResource{}
	_ resource.ResourceWithConfigure = &organisationResource{}
)

// NewOrganisationResource is a helper function to simplify the provider implementation.
func NewOrganisationResource() resource.Resource {
	return &organisationResource{}
}

// organisationResource is the resource implementation.
type organisationResource struct {
	client qernalclient.QernalAPIClient
}

func (r *organisationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *organisationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organisation"
}

// Schema defines the schema for the resource.
func (r *organisationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Computed: true,
				Required: false,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"date": schema.SingleNestedAttribute{
				Required: false,
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
func (r *organisationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan organisationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new organisation
	org, _, err := r.client.OrganisationsAPI.OrganisationsCreate(ctx).OrganisationBody(openapiclient.OrganisationBody{
		Name: plan.Name.ValueString(),
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating organisation",
			"Could not create organisation, unexpected error: "+err.Error())
		return
	}

	plan.ID = types.StringValue(org.Id)
	plan.Name = types.StringValue(org.Name)
	plan.UserID = types.StringValue(org.UserId)
	date := resourceDate{
		CreatedAt: org.Date.CreatedAt,
		UpdatedAt: org.Date.UpdatedAt,
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
func (r *organisationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state organisationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed organisation value from qernal
	org, _, err := r.client.OrganisationsAPI.OrganisationsGet(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading organisation",
			"Could not read organisation ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(org.Name)
	state.UserID = types.StringValue(org.UserId)
	date := resourceDate{
		CreatedAt: org.Date.CreatedAt,
		UpdatedAt: org.Date.UpdatedAt,
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
func (r *organisationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan organisationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing organisation
	_, _, err := r.client.OrganisationsAPI.OrganisationsUpdate(ctx, plan.ID.ValueString()).OrganisationBody(
		openapiclient.OrganisationBody{
			Name: plan.Name.ValueString(),
		}).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating organisation",
			"Could not update organisation, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch updated organisation
	org, _, err := r.client.OrganisationsAPI.OrganisationsGet(ctx, plan.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading organisation",
			"Could not read organisation ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.Name = types.StringValue(org.Name)
	plan.UserID = types.StringValue(org.UserId)
	date := resourceDate{
		CreatedAt: org.Date.CreatedAt,
		UpdatedAt: org.Date.UpdatedAt,
	}
	plan.Date = date.GetDateObject()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *organisationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state organisationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	_, _, err := r.client.OrganisationsAPI.OrganisationsDelete(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting organisation",
			"Could not delete organisation, unexpected error: "+err.Error(),
		)
		return
	}
}

// organisationResourceModel maps the resource schema data.
type organisationResourceModel struct {
	ID     types.String          `tfsdk:"id"`
	Name   types.String          `tfsdk:"name"`
	UserID types.String          `tfsdk:"user_id"`
	Date   basetypes.ObjectValue `tfsdk:"date"`
}

type resourceDate struct {
	CreatedAt string `tfsdk:"created_at"`
	UpdatedAt string `tfsdk:"updated_at"`
}

func (r resourceDate) GetDateObject() basetypes.ObjectValue {
	elementTypes := map[string]attr.Type{
		"created_at": types.StringType,
		"updated_at": types.StringType,
	}
	elements := map[string]attr.Value{
		"created_at": types.StringValue(r.CreatedAt),
		"updated_at": types.StringValue(r.UpdatedAt),
	}
	objectValue, _ := types.ObjectValue(elementTypes, elements)

	return objectValue
}
