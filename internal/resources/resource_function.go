package resources

import (
	"context"
	"fmt"
	qernalclient "terraform-provider-qernal/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	openapiclient "github.com/qernal/openapi-chaos-go-client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &FunctionResource{}
	_ resource.ResourceWithConfigure = &FunctionResource{}
)

// NewFunctionResource is a helper function to simplify the provider implementation.
func NewFunctionResource() resource.Resource {
	return &FunctionResource{}
}

// FunctionResource is the resource implementation.
type FunctionResource struct {
	client qernalclient.QernalAPIClient
}

func (r *FunctionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *FunctionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (r *FunctionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: false,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Unique identifier for the function, assigned automatically upon creation.",
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "ID of the project where the function will be deployed.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Name of the function.",
			},
			"version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Version of the function.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A brief description of the function.",
			},
			"image": schema.StringAttribute{
				Required:    true,
				Description: "Container image for the function.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of the function (e.g., HTTP, Event-driven).",
			},
			"size": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Size configuration for the function, specifying CPU and memory resources.",
				Attributes: map[string]schema.Attribute{
					"cpu": schema.Int64Attribute{
						Required:    true,
						Description: "Amount of CPU allocated to the function.",
					},
					"memory": schema.Int64Attribute{
						Required:    true,
						Description: "Amount of memory allocated to the function.",
					},
				},
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "Port on which the function will listen for incoming requests.",
			},
			"routes": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of routes that define the function's endpoints.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Required:    true,
							Description: "Path of the route.",
						},
						"methods": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "HTTP methods supported by the route (e.g., GET, POST).",
						},
						"weight": schema.Int64Attribute{
							Required:    true,
							Description: "Weight of the route for load balancing.",
						},
					},
				},
			},
			"scaling": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Scaling configuration for the function.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:    true,
						Description: "Type of scaling (e.g., automatic, manual).",
					},
					"low": schema.Int64Attribute{
						Required:    true,
						Description: "Lower bound for scaling.",
					},
					"high": schema.Int64Attribute{
						Required:    true,
						Description: "Upper bound for scaling.",
					},
				},
			},
			"deployments": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of deployments for the function, specifying locations and replicas.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Deployment location details.",
							Attributes: map[string]schema.Attribute{
								"provider_id": schema.StringAttribute{
									Required:    true,
									Description: "ID of the cloud provider.",
								},
								"continent": schema.StringAttribute{
									Required:    true,
									Description: "Continent where the deployment is located.",
								},
								"country": schema.StringAttribute{
									Required:    true,
									Description: "Country where the deployment is located.",
								},
								"city": schema.StringAttribute{
									Required:    true,
									Description: "City where the deployment is located.",
								},
							},
						},
						"replicas": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Replica configuration for the deployment.",
							Attributes: map[string]schema.Attribute{
								"min": schema.Int64Attribute{
									Required:    true,
									Description: "Minimum number of replicas.",
								},
								"max": schema.Int64Attribute{
									Required:    true,
									Description: "Maximum number of replicas.",
								},
								"affinity": schema.SingleNestedAttribute{
									Required:    true,
									Description: "Affinity settings for replicas.",
									Attributes: map[string]schema.Attribute{
										"cluster": schema.BoolAttribute{
											Required:    true,
											Description: "Indicates if replicas should be clustered.",
										},
										"cloud": schema.BoolAttribute{
											Required:    true,
											Description: "Indicates if replicas should be spread across multiple clouds.",
										},
									},
								},
							},
						},
					},
				},
			},
			"secrets": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of secrets used by the function.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the secret.",
						},
						"reference": schema.StringAttribute{
							Required:    true,
							Description: "Reference to the secret's value.",
						},
					},
				},
			},
			"compliance": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Compliance standards the function adheres to, one of soc or ipv6 ",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *FunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan FunctionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	functionSize := openapiclient.NewFunctionSize(int32(plan.Size.CPU.ValueInt64()), int32(plan.Size.Memory.ValueInt64()))
	functionRoutes := RoutesToOpenAPI(plan.Routes)
	functionDeployment := DeploymentsToOepnAPI(plan.Deployments)
	functionSecrets := SecretsToOpenAPI(plan.Secrets)
	functionCompliance := ComplianceToOpenAPI(plan.Compliance)

	// Create new Function
	function, httpRes, err := r.client.FunctionsAPI.ProjectsFunctionsCreate(ctx, plan.ProjectID.ValueString()).FunctionBody(openapiclient.FunctionBody{
		ProjectId:   plan.ProjectID.ValueString(),
		Version:     plan.Version.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Image:       plan.Image.ValueString(),
		Type:        openapiclient.FunctionType(plan.FunctionType.ValueString()),
		Size:        *functionSize,
		Port:        int32(plan.Port.ValueInt64()),
		Routes:      functionRoutes,
		Scaling: openapiclient.FunctionScaling{
			Type: plan.Scaling.Type.ValueString(),
			Low:  int32(plan.Scaling.Low.ValueInt64()),
			High: int32(plan.Scaling.High.ValueInt64()),
		},
		Deployments: functionDeployment,
		Secrets:     functionSecrets,
		Compliance:  functionCompliance,
	}).Execute()
	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)
		fmt.Println(err)
		resp.Diagnostics.AddError(
			"Error creating Function",
			"Could not create Function, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	plan.ID = types.StringValue(function.Id)
	plan.ProjectID = types.StringValue(function.ProjectId)
	plan.Version = types.StringValue(function.Version)
	plan.Name = types.StringValue(function.Name)
	plan.Description = types.StringValue(function.Description)
	plan.Image = types.StringValue(function.Image)
	plan.FunctionType = types.StringValue(string(function.Type))
	plan.Size = Size{
		CPU:    types.Int64Value(int64(function.Size.Cpu)),
		Memory: types.Int64Value(int64(function.Size.Memory)),
	}

	plan.Port = types.Int64Value(int64(function.Port))

	plan.Routes = OpenAPIToRoutes(function.Routes)
	plan.Scaling = Scaling{
		Type: types.StringValue(function.Scaling.Type),
		Low:  types.Int64Value(int64(function.Scaling.Low)),
		High: types.Int64Value(int64(function.Scaling.High)),
	}

	plan.Deployments = OpenAPIDeploymentsToDeployments(function.Deployments)

	plan.Secrets = FunctionEnvsToSecrets(function.Secrets)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state FunctionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed Function value from qernal
	function, _, err := r.client.FunctionsAPI.FunctionsGet(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Function",
			"Could not read Function Name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(function.Id)
	state.ProjectID = types.StringValue(function.ProjectId)
	state.Version = types.StringValue(function.Version)
	state.Name = types.StringValue(function.Name)
	state.Description = types.StringValue(function.Description)
	state.Image = types.StringValue(function.Image)
	state.FunctionType = types.StringValue(string(function.Type))
	state.Size = Size{
		CPU:    types.Int64Value(int64(function.Size.Cpu)),
		Memory: types.Int64Value(int64(function.Size.Memory)),
	}

	state.Port = types.Int64Value(int64(function.Port))

	state.Routes = OpenAPIToRoutes(function.Routes)
	state.Scaling = Scaling{
		Type: types.StringValue(function.Scaling.Type),
	}

	state.Deployments = OpenAPIDeploymentsToDeployments(function.Deployments)

	state.Compliance = OpenAPIToCompliance(function.Compliance)

	state.Secrets = FunctionEnvsToSecrets(function.Secrets)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *FunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan FunctionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing Function
	functionSize := openapiclient.NewFunctionSize(int32(plan.Size.CPU.ValueInt64()), int32(plan.Size.Memory.ValueInt64()))
	functionRoutes := RoutesToOpenAPI(plan.Routes)

	//TODO: Needs work
	//functionDeployment := DeploymentsToOepnAPI(plan.Deployments)
	functionSecrets := SecretsToOpenAPI(plan.Secrets)
	functionCompliance := ComplianceToOpenAPI(plan.Compliance)

	_, httpRes, err := r.client.FunctionsAPI.FunctionsUpdate(ctx, plan.ID.ValueString()).Function(openapiclient.Function{
		ProjectId:   plan.ID.ValueString(),
		Version:     plan.Version.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Name.ValueString(),
		Image:       plan.Image.ValueString(),
		Type:        openapiclient.FunctionType(plan.FunctionType.ValueString()),
		Size:        *functionSize,
		Port:        int32(plan.Port.ValueInt64()),
		Routes:      functionRoutes,
		Scaling: openapiclient.FunctionScaling{
			Type: plan.Scaling.Type.String(),
			Low:  int32(plan.Scaling.Low.ValueInt64()),
			High: int32(plan.Scaling.High.ValueInt64()),
		},
		Deployments: make([]openapiclient.FunctionDeployment, 0),
		Secrets:     functionSecrets,
		Compliance:  functionCompliance,
	}).Execute()
	if err != nil {
		resData, _ := qernalclient.ParseResponseData(httpRes)
		resp.Diagnostics.AddError(
			"Error updating Function",
			"Could not update Function, unexpected error: "+err.Error()+" with"+fmt.Sprintf(", detail: %v", resData))
		return
	}

	// Fetch updated Function
	function, _, err := r.client.FunctionsAPI.FunctionsGet(ctx, plan.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Function",
			"Could not read Function Name "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items
	plan.ID = types.StringValue(function.Id)
	plan.ProjectID = types.StringValue(function.ProjectId)
	plan.Version = types.StringValue(function.Version)
	plan.Name = types.StringValue(function.Name)
	plan.Description = types.StringValue(function.Description)
	plan.Image = types.StringValue(function.Image)
	plan.FunctionType = types.StringValue(string(function.Type))
	plan.Size = Size{
		CPU:    types.Int64Value(int64(function.Size.Cpu)),
		Memory: types.Int64Value(int64(function.Size.Memory)),
	}

	plan.Port = types.Int64Value(int64(function.Port))

	plan.Routes = OpenAPIToRoutes(function.Routes)
	plan.Scaling = Scaling{
		Type: types.StringValue(function.Scaling.Type),
	}

	plan.Deployments = OpenAPIDeploymentsToDeployments(function.Deployments)

	plan.Compliance = OpenAPIToCompliance(function.Compliance)

	plan.Secrets = FunctionEnvsToSecrets(function.Secrets)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state FunctionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing Function
	_, _, err := r.client.FunctionsAPI.FunctionsDelete(ctx, state.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Function",
			"Could not delete Function, unexpected error: "+err.Error(),
		)
		return
	}
}

// FunctionResourceModel maps the resource schema data.
type FunctionResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	ProjectID    types.String   `tfsdk:"project_id"`
	Name         types.String   `tfsdk:"name"`
	Version      types.String   `tfsdk:"version"`
	Description  types.String   `tfsdk:"description"`
	Image        types.String   `tfsdk:"image"`
	FunctionType types.String   `tfsdk:"type"`
	Size         Size           `tfsdk:"size"`
	Port         types.Int64    `tfsdk:"port"`
	Routes       []Route        `tfsdk:"routes"`
	Scaling      Scaling        `tfsdk:"scaling"`
	Deployments  []Deployment   `tfsdk:"deployments"`
	Secrets      []Secret       `tfsdk:"secrets"`
	Compliance   []types.String `tfsdk:"compliance"`
}

// Size represents the size configuration for the function
type Size struct {
	CPU    types.Int64 `tfsdk:"cpu"`
	Memory types.Int64 `tfsdk:"memory"`
}

// Route represents a route configuration for the function
type Route struct {
	Path    types.String   `tfsdk:"path"`
	Methods []types.String `tfsdk:"methods"`
	Weight  types.Int64    `tfsdk:"weight"`
}

// Scaling represents the scaling configuration for the function
type Scaling struct {
	Type types.String `tfsdk:"type"`
	Low  types.Int64  `tfsdk:"low"`
	High types.Int64  `tfsdk:"high"`
}

// Deployment represents a deployment configuration for the function
type Deployment struct {
	Location Location `tfsdk:"location"`
	Replicas Replicas `tfsdk:"replicas"`
}

// Location represents the location details for a deployment
type Location struct {
	ProviderID types.String `tfsdk:"provider_id"`
	Continent  types.String `tfsdk:"continent"`
	Country    types.String `tfsdk:"country"`
	City       types.String `tfsdk:"city"`
}

// Replicas represents the replicas configuration for a deployment
type Replicas struct {
	Min      types.Int64 `tfsdk:"min"`
	Max      types.Int64 `tfsdk:"max"`
	Affinity Affinity    `tfsdk:"affinity"`
}

// Affinity represents the affinity configuration for replicas
type Affinity struct {
	Cluster types.Bool `tfsdk:"cluster"`
	Cloud   types.Bool `tfsdk:"cloud"`
}

// Secret represents a secret configuration for the function
type Secret struct {
	Name      types.String `tfsdk:"name"`
	Reference types.String `tfsdk:"reference"`
}

func RoutesToOpenAPI(routes []Route) []openapiclient.FunctionRoute {
	var openAPIRoutes []openapiclient.FunctionRoute

	for _, route := range routes {
		// Convert types.String to string and types.Int64 to int32
		path := route.Path.ValueString()
		methods := make([]string, len(route.Methods))
		for i, method := range route.Methods {
			methods[i] = method.ValueString()
		}
		weight := int32(route.Weight.ValueInt64())

		// Call openapiclient.NewFunctionRoute
		openAPIRoute := openapiclient.NewFunctionRoute(path, methods, weight)

		// Append the result to the list
		openAPIRoutes = append(openAPIRoutes, *openAPIRoute)
	}

	return openAPIRoutes
}

func DeploymentsToOepnAPI(deployments []Deployment) []openapiclient.FunctionDeploymentBody {
	var openAPIDeploymentBody []openapiclient.FunctionDeploymentBody

	for _, deploy := range deployments {
		openAPIdeploy := openapiclient.FunctionDeploymentBody{
			Location: openapiclient.Location{
				ProviderId: deploy.Location.ProviderID.ValueString(),
				Continent:  deploy.Location.Continent.ValueStringPointer(),
				Country:    deploy.Location.Country.ValueStringPointer(),
				City:       deploy.Location.City.ValueStringPointer(),
			},
			Replicas: openapiclient.FunctionReplicas{
				Min: int32(deploy.Replicas.Min.ValueInt64()),
				Max: int32(deploy.Replicas.Max.ValueInt64()),
				Affinity: openapiclient.FunctionReplicasAffinity{
					Cluster: deploy.Replicas.Affinity.Cluster.ValueBool(),
					Cloud:   deploy.Replicas.Affinity.Cloud.ValueBool(),
				},
			},
		}

		openAPIDeploymentBody = append(openAPIDeploymentBody, openAPIdeploy)
	}

	return openAPIDeploymentBody
}

// OpenAPIDeploymentsToDeployments converts a list of openapiclient.FunctionDeploymentBody to a list of Deployment structs
func OpenAPIDeploymentsToDeployments(openAPIDeployments []openapiclient.FunctionDeployment) []Deployment {
	var deployments []Deployment

	for _, openAPIDeployment := range openAPIDeployments {
		// Convert Location fields
		providerID := types.StringValue(openAPIDeployment.Location.ProviderId)
		continent := types.StringPointerValue(openAPIDeployment.Location.Continent)
		country := types.StringPointerValue(openAPIDeployment.Location.Country)
		city := types.StringPointerValue(openAPIDeployment.Location.City)

		// Convert Replicas fields
		min := types.Int64Value(int64(openAPIDeployment.Replicas.Min))
		max := types.Int64Value(int64(openAPIDeployment.Replicas.Max))
		cluster := types.BoolValue(openAPIDeployment.Replicas.Affinity.Cluster)
		cloud := types.BoolValue(openAPIDeployment.Replicas.Affinity.Cloud)

		// Create Location and Replicas for Deployment
		location := Location{
			ProviderID: providerID,
			Continent:  continent,
			Country:    country,
			City:       city,
		}

		replicas := Replicas{
			Min: min,
			Max: max,
			Affinity: Affinity{
				Cluster: cluster,
				Cloud:   cloud,
			},
		}

		// Create a Deployment instance
		deployment := Deployment{
			Location: location,
			Replicas: replicas,
		}

		// Append the result to the list
		deployments = append(deployments, deployment)
	}

	return deployments
}

func SecretsToOpenAPI(secrets []Secret) []openapiclient.FunctionEnv {

	var openAPIFunctionEnv []openapiclient.FunctionEnv
	if len(openAPIFunctionEnv) <= 0 {
		return []openapiclient.FunctionEnv{}
	}

	for _, env := range secrets {

		funcEnv := openapiclient.NewFunctionEnv(env.Name.String(), env.Reference.ValueString())

		openAPIFunctionEnv = append(openAPIFunctionEnv, *funcEnv)
	}

	return openAPIFunctionEnv
}

func ComplianceToOpenAPI(compliance []types.String) []openapiclient.FunctionCompliance {
	var openAPICompliance []openapiclient.FunctionCompliance
	for _, comp := range compliance {
		openAPICompliance = append(openAPICompliance, openapiclient.FunctionCompliance(comp.ValueString()))
	}
	return openAPICompliance
}

// OpenAPIToCompliance converts a slice of openapiclient.FunctionCompliance to a slice of types.String
func OpenAPIToCompliance(openAPICompliance []openapiclient.FunctionCompliance) []types.String {

	// Initialize an empty slice of types.String with the same length as openAPICompliance
	compliance := make([]types.String, len(openAPICompliance))

	// Iterate over each openAPICompliance
	for _, comp := range openAPICompliance {
		// Convert each openapiclient.FunctionCompliance to types.String
		compliance = append(compliance, types.StringValue(string(comp)))
	}

	return compliance
}

func OpenAPIToRoutes(openAPIRoutes []openapiclient.FunctionRoute) []Route {
	var routes []Route

	for _, openAPIRoute := range openAPIRoutes {
		// Convert string to types.String and int32 to types.Int64
		path := types.StringValue(openAPIRoute.Path)
		methods := make([]types.String, len(openAPIRoute.Methods))
		for i, method := range openAPIRoute.Methods {
			methods[i] = types.StringValue(method)
		}
		weight := types.Int64Value(int64(openAPIRoute.Weight))

		// Create a Route instance
		route := Route{
			Path:    path,
			Methods: methods,
			Weight:  weight,
		}

		// Append the result to the list
		routes = append(routes, route)
	}

	return routes
}

// FunctionEnvsToSecrets converts a slice of FunctionEnv to a slice of Secret
func FunctionEnvsToSecrets(functionEnvs []openapiclient.FunctionEnv) []Secret {
	if len(functionEnvs) < 1 {
		return []Secret{}
	}
	// Initialize an empty slice of Secret with the same length as functionEnvs
	secrets := make([]Secret, len(functionEnvs))
	for _, functionEnv := range functionEnvs {
		// Convert Name and Reference to types.String
		name := types.StringValue(functionEnv.Name)
		reference := types.StringValue(functionEnv.Reference)

		// Create a Secret instance
		secret := Secret{
			Name:      name,
			Reference: reference,
		}

		secrets = append(secrets, secret)
	}
	return secrets
}
