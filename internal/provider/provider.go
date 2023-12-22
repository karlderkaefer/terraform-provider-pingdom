// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &PingdomProvider{}

// PingdomProvider defines the provider implementation.
type PingdomProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PingdomProviderModel describes the provider data model.
type PingdomProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	ApiToken types.String `tfsdk:"api_token"`
}

func (p *PingdomProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pingdom"
	resp.Version = p.version
}

func (p *PingdomProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Pingdom API endpoint",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Pingdom API token for authentication",

				Required:  true,
				Sensitive: true, // Mark as sensitive to prevent logging
			},
		},
	}
}

func (p *PingdomProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PingdomProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use api_token from configuration, or fallback to environment variable if not provided
	apiToken := data.ApiToken.ValueString()
	if apiToken == "" {
		apiToken = os.Getenv("PINGDOM_API_TOKEN")
	}

	if apiToken == "" {
		resp.Diagnostics.AddError("Missing API Token", "The 'api_token' attribute is required but was not provided, and environment variable 'PINGDOM_API_TOKEN' is not set.")
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	bearerTokenProvider, err := securityprovider.NewSecurityProviderBearerToken(apiToken)
	if err != nil {
		tflog.Error(ctx, "Error creating Pingdom client: "+err.Error())
	}
	resp.DataSourceData = bearerTokenProvider
	resp.ResourceData = bearerTokenProvider
}

func (p *PingdomProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *PingdomProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTransactionChecksDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PingdomProvider{
			version: version,
		}
	}
}
