package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	qernalclient "terraform-provider-qernal/internal/client"
	qernalresource "terraform-provider-qernal/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// qernalProviderModel maps provider schema data to a Go type.
type qernalProviderModel struct {
	Token     types.String `tfsdk:"token"`
	HostChaos types.String `tfsdk:"host_chaos"`
	HostHydra types.String `tfsdk:"host_hydra"`
}

var (
	_ provider.Provider = &qernalProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &qernalProvider{
			version: version,
		}
	}
}

type qernalProvider struct {
	version string
}

func (p *qernalProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "qernal"
	resp.Version = p.version
}

func (p *qernalProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host_chaos": schema.StringAttribute{
				Optional:    true,
				Description: "The endpoint of Qernal Cloud API",
			},
			"host_hydra": schema.StringAttribute{
				Optional:    true,
				Description: "The endpoint of OAuth 2 server",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Description: "The token use to authenticate with the qernal platform, with format: client_id@clien_secret",
			},
		},
	}
}

func (p *qernalProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config qernalProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown token",
			"The provider cannot create the Qernal API client as there is an unknown configuration value for the token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the QERNAL_TOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	hostChaos := os.Getenv("QERNAL_HOST_CHAOS")
	if hostChaos == "" {
		hostChaos = "https://chaos.qernal.com"
	}

	hostHydra := os.Getenv("QERNAL_HOST_HYDRA")
	if hostHydra == "" {
		hostHydra = "https://hydra.qernal.com"
	}
	token := os.Getenv("QERNAL_TOKEN")

	if !config.HostChaos.IsNull() {
		hostChaos = config.HostChaos.ValueString()
	}

	if !config.HostHydra.IsNull() {
		hostHydra = config.HostHydra.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing token",
			"The provider cannot create the Qernal API client as there is a missing or empty value for the Qernal token. "+
				"Set the password value in the configuration or use the QERNAL_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Qernal client using the configuration values
	client, err := qernalclient.New(ctx, hostHydra, hostChaos, token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Qernal API Client",
			"An unexpected error occurred when creating the Qernal API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Qernal Client Error: "+err.Error(),
		)
		return
	}

	// Make the Qernal client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *qernalProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *qernalProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		qernalresource.NewOrganisationResource,
		qernalresource.NewProjectResource,
		qernalresource.NewSecretResource,
		qernalresource.NewTokenResource,
		qernalresource.NewHostResource,
	}
}
