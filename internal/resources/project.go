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
	qernalclient "qernal-terraform-provider/internal/client"
)

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &projectResource{}
	_ resource.ResourceWithConfigure = &projectResource{}
)

// NewProjectResource is a helper function to simplify the provider implementation.
func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// ProjectResource is the resource implementation.
type projectResource struct {
	client qernalclient.QernalAPIClient
}

func (r *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the resource.
func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
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
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan ProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Project
	org, _, err := r.client.ProjectsAPI.ProjectsCreate(ctx).ProjectBody(openapiclient.ProjectBody{
		Name:  plan.Name.ValueString(),
		OrgId: plan.OrgID.ValueString(),
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Project",
			"Could not create Project, unexpected error: "+err.Error())
		return
	}

	plan.ID = types.StringValue(org.Id)
	plan.Name = types.StringValue(org.Name)
	plan.OrgID = types.StringValue(org.OrgId)
	date, _ := types.ObjectValue(
		map[string]attr.Type{
			"created_at": types.StringType,
			"updated_at": types.StringType,
		},
		map[string]attr.Value{
			"created_at": types.StringValue(org.Date.CreatedAt),
			"updated_at": types.StringValue(org.Date.UpdatedAt),
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
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed Project value from qernal
	org, _, err := r.client.ProjectsAPI.ProjectsGet(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Project",
			"Could not read Project ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(org.Name)
	state.OrgID = types.StringValue(org.OrgId)
	state.Date, _ = types.ObjectValue(
		map[string]attr.Type{
			"created_at": types.StringType,
			"updated_at": types.StringType,
		},
		map[string]attr.Value{
			"created_at": types.StringValue(org.Date.CreatedAt),
			"updated_at": types.StringValue(org.Date.UpdatedAt),
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
func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan ProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing Project
	_, _, err := r.client.ProjectsAPI.ProjectsUpdate(ctx, plan.ID.ValueString()).ProjectBodyPatch(
		openapiclient.ProjectBodyPatch{
			Name:  plan.Name.ValueStringPointer(),
			OrgId: plan.OrgID.ValueStringPointer(),
		}).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Project",
			"Could not update Project, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch updated Project
	org, _, err := r.client.ProjectsAPI.ProjectsGet(ctx, plan.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Project",
			"Could not read Project ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.Name = types.StringValue(org.Name)
	plan.OrgID = types.StringValue(org.OrgId)
	plan.Date, _ = types.ObjectValue(
		map[string]attr.Type{
			"created_at": types.StringType,
			"updated_at": types.StringType,
		},
		map[string]attr.Value{
			"created_at": types.StringValue(org.Date.CreatedAt),
			"updated_at": types.StringValue(org.Date.UpdatedAt),
		},
	)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	_, _, err := r.client.ProjectsAPI.ProjectsDelete(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Project",
			"Could not delete Project, unexpected error: "+err.Error(),
		)
		return
	}
}

// ProjectResourceModel maps the resource schema data.
type ProjectResourceModel struct {
	ID    types.String          `tfsdk:"id"`
	Name  types.String          `tfsdk:"name"`
	OrgID types.String          `tfsdk:"org_id"`
	Date  basetypes.ObjectValue `tfsdk:"date"`
}

type ProjectDate struct {
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}
