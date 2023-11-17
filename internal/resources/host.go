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
	pkgtypes "terraform-provider-qernal/pkg/types"
)

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &hostResource{}
	_ resource.ResourceWithConfigure = &hostResource{}
)

// NewHostResource is a helper function to simplify the provider implementation.
func NewHostResource() resource.Resource {
	return &hostResource{}
}

// hostResource is the resource implementation.
type hostResource struct {
	client qernalclient.QernalAPIClient
}

func (r *hostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

// Schema defines the schema for the resource.
func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: false,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
			"certificate": schema.StringAttribute{
				Required: true,
			},
			"disabled": schema.BoolAttribute{
				Required: true,
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

// Create creates the resource and sets the initial Terraform state.
func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan hostResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new host

	host, httpRes, err := r.client.HostsAPI.ProjectsHostsCreate(ctx, plan.ProjectID.ValueString()).HostBody(openapiclient.HostBody{
		Host:        plan.Name.ValueString(),
		Certificate: plan.Certificate.ValueString(),
		Disabled:    plan.Disabled.ValueBool(),
	}).Execute()
	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error creating Host",
			"Could not create Host, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	plan.ID = types.StringValue(host.Id)
	plan.Name = types.StringValue(host.Host)
	plan.ProjectID = types.StringValue(host.ProjectId)
	plan.Certificate = types.StringValue(pkgtypes.StringValueFromPointer(host.Certificate))
	plan.ReadOnly = types.BoolValue(host.ReadOnly)
	//plan.Disabled = types.BoolValue(host.Disabled)
	plan.TxtVerification = types.StringValue(host.TxtVerification)
	plan.VerifiedAt = types.StringValue(pkgtypes.StringValueFromPointer(host.VerifiedAt))
	plan.VerificationStatus = types.StringValue(string(host.VerificationStatus))

	date := resourceDate{
		CreatedAt: host.Date.CreatedAt,
		UpdatedAt: host.Date.UpdatedAt,
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
func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed host value from qernal
	host, _, err := r.client.HostsAPI.ProjectsHostsGet(ctx, state.ProjectID.ValueString(), state.Name.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Host",
			"Could not read Host Name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(host.Host)
	state.ProjectID = types.StringValue(host.ProjectId)
	state.Certificate = types.StringValue(pkgtypes.StringValueFromPointer(host.Certificate))
	state.ReadOnly = types.BoolValue(host.ReadOnly)
	state.Disabled = types.BoolValue(host.Disabled)
	state.TxtVerification = types.StringValue(host.TxtVerification)
	state.VerifiedAt = types.StringValue(pkgtypes.StringValueFromPointer(host.VerifiedAt))
	state.VerificationStatus = types.StringValue(string(host.VerificationStatus))

	date := resourceDate{
		CreatedAt: host.Date.CreatedAt,
		UpdatedAt: host.Date.UpdatedAt,
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
func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan hostResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing host

	_, httpRes, err := r.client.HostsAPI.ProjectsHostsUpdate(ctx, plan.ProjectID.ValueString(), plan.Name.ValueString()).HostBodyPatch(openapiclient.HostBodyPatch{
		Certificate: plan.Certificate.ValueStringPointer(),
		Disabled:    plan.Disabled.ValueBoolPointer(),
	}).Execute()
	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error updating Host",
			"Could not update Host, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	// Fetch updated Host
	host, _, err := r.client.HostsAPI.ProjectsHostsGet(ctx, plan.ProjectID.ValueString(), plan.Name.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Host",
			"Could not read Host Name "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp

	plan.Name = types.StringValue(host.Host)
	plan.ProjectID = types.StringValue(host.ProjectId)
	plan.Certificate = types.StringValue(pkgtypes.StringValueFromPointer(host.Certificate))
	plan.ReadOnly = types.BoolValue(host.ReadOnly)
	plan.Disabled = types.BoolValue(host.Disabled)
	plan.TxtVerification = types.StringValue(host.TxtVerification)
	plan.VerifiedAt = types.StringValue(pkgtypes.StringValueFromPointer(host.VerifiedAt))
	plan.VerificationStatus = types.StringValue(string(host.VerificationStatus))

	date := resourceDate{
		CreatedAt: host.Date.CreatedAt,
		UpdatedAt: host.Date.UpdatedAt,
	}
	plan.Date = date.GetDateObject()

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing host
	_, _, err := r.client.HostsAPI.ProjectsHostsDelete(ctx, state.ProjectID.ValueString(), state.Name.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting host",
			"Could not delete host, unexpected error: "+err.Error(),
		)
		return
	}
}

// hostResourceModel maps the resource schema data.
type hostResourceModel struct {
	ID                 types.String          `tfsdk:"id"`
	ProjectID          types.String          `tfsdk:"project_id"`
	Name               types.String          `tfsdk:"name"`
	Certificate        types.String          `tfsdk:"certificate"`
	ReadOnly           types.Bool            `tfsdk:"read_only"`
	Disabled           types.Bool            `tfsdk:"disabled"`
	TxtVerification    types.String          `tfsdk:"txt_verification"`
	VerifiedAt         types.String          `tfsdk:"verified_at"`
	VerificationStatus types.String          `tfsdk:"verification_status"`
	Date               basetypes.ObjectValue `tfsdk:"date"`
}
