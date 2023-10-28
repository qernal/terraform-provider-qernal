package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	qernalclient "qernal-terraform-provider/internal/client"
	qernalresource "qernal-terraform-provider/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// qernalProviderModel maps provider schema data to a Go type.
type qernalProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
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
			"host": schema.StringAttribute{
				Optional: true,
			},
			"token": schema.StringAttribute{
				Required: true,
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

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown host",
			"The provider cannot create the Qernal API client as there is an unknown configuration value for the host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the QERNAL_HOST environment variable.",
		)
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

	host := os.Getenv("QERNAL_HOST")
	if host == "" {
		host = "https://chaos.qernal.com/v1"
	}
	token := os.Getenv("QERNAL_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Qernal API Host",
			"The provider cannot create the Qernal API client as there is a missing or empty value for the Qernal API host. "+
				"Set the host value in the configuration or use the QERNAL_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
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
	client, err := qernalclient.New(host, token)
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
	}
}
