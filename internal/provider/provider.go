package provider

import (
	"context"
	shopware_sdk "github.com/friendsofshopware/go-shopware-admin-api-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ShopwareProvider satisfies various provider interfaces.
var _ provider.Provider = &ShopwareProvider{}

// ShopwareProvider defines the provider implementation.
type ShopwareProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ShopwareProviderModel describes the provider data model.
type ShopwareProviderModel struct {
	URL           types.String `tfsdk:"url"`
	ClientId      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	AdminUsername types.String `tfsdk:"admin_username"`
	AdminPassword types.String `tfsdk:"admin_password"`
}

func (p *ShopwareProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "shopware"
	resp.Version = p.version
}

func (p *ShopwareProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of Shopware instance",
				Required:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Client ID of Integration",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Client Secret of Integration",
				Optional:            true,
			},
			"admin_username": schema.StringAttribute{
				MarkdownDescription: "Client ID of Integration",
				Optional:            true,
			},
			"admin_password": schema.StringAttribute{
				MarkdownDescription: "Client Secret of Integration",
				Optional:            true,
			},
		},
	}
}

func (p *ShopwareProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ShopwareProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var creds shopware_sdk.OAuthCredentials

	if data.ClientId.IsNull() {
		creds = shopware_sdk.NewPasswordCredentials(data.AdminUsername.ValueString(), data.AdminPassword.ValueString(), []string{"write"})
	} else {
		creds = shopware_sdk.NewIntegrationCredentials(data.ClientId.ValueString(), data.ClientSecret.ValueString(), []string{"write"})
	}

	client, err := shopware_sdk.NewApiClient(context.Background(), data.URL.ValueString(), creds, nil)

	if err != nil {
		resp.Diagnostics.AddError("Cannot authenticate to Shop", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ShopwareProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSystemConfigResource,
		NewShippingMethodResource,
		NewRuleResource,
	}
}

func (p *ShopwareProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ShopwareProvider{
			version: version,
		}
	}
}
