package provider

import (
	"context"
	"fmt"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/karlderkaefer/pingdom-golang-client/pkg/pingdom/client"
	"github.com/karlderkaefer/pingdom-golang-client/pkg/pingdom/client/tmschecks"
)

var _ datasource.DataSource = (*transactionChecksDataSource)(nil)

func NewTransactionChecksDataSource() datasource.DataSource {
	return &transactionChecksDataSource{}
}

type transactionChecksDataSource struct {
	authenticationProvider *securityprovider.SecurityProviderBearerToken
	client                 *tmschecks.ClientWithResponses
}

func (d *transactionChecksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transaction_checks"
}

func (d *transactionChecksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (d *transactionChecksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TransactionChecksModel
	tflog.Info(ctx, fmt.Sprintf("Reading Pingdom transaction checks"))
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		//return
	}

	// Read API call logic
	pingdomResp, err := d.client.GetAllChecksWithResponse(ctx, &tmschecks.GetAllChecksParams{})

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading Pingdom transaction checks: %s", err.Error()))
		return
	}
	if pingdomResp.JSON200 == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading Pingdom transaction checks"), err.Error())
		return
	}
	// for _, p := range *pingdomResp.JSON200.Check {
	// 	tflog.Info(ctx, fmt.Sprintf("Got response %v", &p.Name))
	// }

	// Example data value setting
	data.Fill(ctx, *pingdomResp)
	tflog.Info(ctx, fmt.Sprintf("Got response %+v", data))
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *transactionChecksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	authenticationProvider, ok := req.ProviderData.(*securityprovider.SecurityProviderBearerToken)
	if !ok {
		tflog.Error(ctx, "Unable to cast provider data to Pingdom client")
		return
	}
	client, err := tmschecks.NewClientWithResponses(
		client.DefaultBaseURL,
		tmschecks.WithRequestEditorFn(authenticationProvider.Intercept),
	)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating Pingdom client: %s", err.Error()))
		return
	}
	d.authenticationProvider = authenticationProvider
	d.client = client
}

func (m *ChecksValue) Fill(ctx context.Context, check tmschecks.CheckGeneral) {
	m.Name = types.StringPointerValue(check.Name)
	m.Id = types.Int64PointerValue(check.CheckID)
	m.Region = types.StringPointerValue(check.Region)
	m.Active = types.BoolPointerValue(check.Active)
	m.Tags, _ = types.ListValueFrom(ctx, types.StringType, check.Tags)
}

func (m *TransactionChecksModel) Fill(ctx context.Context, checks tmschecks.GetAllChecksResponse) {
	for _, check := range *checks.JSON200.Check {
		var tfCheck ChecksValue
		tfCheck.Fill(ctx, check)
		tflog.Info(ctx, fmt.Sprintf("\n\nGot parsed check: %+v", tfCheck.state))
		o, _ := tfCheck.ToObjectValue(ctx)
		m.Checks, _ = types.ListValueFrom(ctx, ChecksType{}, o)
		tflog.Info(ctx, fmt.Sprintf("\n\nGot parsed check: %+v", tfCheck))
	}
	m.ExtendedTags = types.BoolValue(false)
	m.Limit = types.StringValue(fmt.Sprintf("%d", *checks.JSON200.Limit))
	m.Offset = types.StringValue(fmt.Sprintf("%d", *checks.JSON200.Offset))

}

