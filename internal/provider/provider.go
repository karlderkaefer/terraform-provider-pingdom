// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	pingdom_client "github.com/karlderkaefer/pingdom-golang-client/pkg/pingdom/client"
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
				Required:            true,
				Sensitive:           true, // Mark as sensitive to prevent logging
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

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := pingdom_client.NewDefaultApiTokenClient(data.ApiToken.ValueString())
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PingdomProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
	}
}

func (p *PingdomProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewChecksDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PingdomProvider{
			version: version,
		}
	}
}
