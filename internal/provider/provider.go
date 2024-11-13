// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	hypervapi "terraform-provider-hyperv/internal/hyper-v-api"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure HyperVProvider satisfies various provider interfaces.
var _ provider.Provider = &HyperVProvider{}
var _ provider.ProviderWithFunctions = &HyperVProvider{}

// HyperVProvider defines the provider implementation.
type HyperVProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HyperVProviderModel describes the provider data model.
type HyperVProviderModel struct {
	Host          types.String `tfsdk:"host"`
	Port          types.Int64  `tfsdk:"port"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	HTTPS         types.Bool   `tfsdk:"https"`
	Insecure      types.Bool   `tfsdk:"insecure"`
	TLSServerName types.String `tfsdk:"tlservername"`
	CACert        types.String `tfsdk:"cacert"`
	CAKey         types.String `tfsdk:"cakey"`
	Cert          types.String `tfsdk:"cert"`
	Timeout       types.Int64  `tfsdk:"timeout"`
}

func (p *HyperVProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hyperv"
	resp.Version = p.version
}

func (p *HyperVProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address of the remote server",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Listening Port of the remote server",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username you wish to authenticate with",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password you wish to authenticate with",
				Required:            true,
			},
			"https": schema.BoolAttribute{
				MarkdownDescription: "Enable connections over HTTPS",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Skip SSL verification",
				Optional:            true,
			},
			"tlservername": schema.StringAttribute{
				MarkdownDescription: "Set to verify the hostname on the returned certificate",
				Optional:            true,
			},
			"cacert": schema.StringAttribute{
				MarkdownDescription: "Set to path of the CACert",
				Optional:            true,
			},
			"cakey": schema.StringAttribute{
				MarkdownDescription: "Set to path of the CAKey",
				Optional:            true,
			},
			"cert": schema.StringAttribute{
				MarkdownDescription: "Set to path of the Cert",
				Optional:            true,
			},

			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Set the timeout to wait for connections to become avalible to hyper-v",
				Optional:            true,
			},
		},
	}
}

func (p *HyperVProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configurating winrm client")
	var tfconfig HyperVProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &tfconfig)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if tfconfig.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Hyper-V Host",
			"The provider cannot create the Hyper-V client as there is an unknown configuration value for the Hyper-V host. "+
				"Either target apply the source of the value first, set the value statically in the configuration",
		)
	}
	if tfconfig.Port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Unknown Hyper-V Port",
			"The provider cannot create the Hyper-V client as there is an unknown configuration value for the Hyper-V port. "+
				"Either target apply the source of the value first, set the value statically in the configuration",
		)
	}
	if tfconfig.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Hyper-V Username",
			"The provider cannot create the Hyper-V client as there is an unknown configuration value for the Hyper-V username. "+
				"Either target apply the source of the value first, set the value statically in the configuration",
		)
	}
	if tfconfig.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Hyper-V Password",
			"The provider cannot create the Hyper-V client as there is an unknown configuration value for the Hyper-V password. "+
				"Either target apply the source of the value first, set the value statically in the configuration",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Setup collecting variables from ENV
	// host := os.Getenv("HYPERV_HOST")
	// port := os.Getenv("HYPERV_PORT")
	// username := os.Getenv("HYPERV_USERNAME")
	// password := os.Getenv("HYPERV_PASSWORD")

	host := tfconfig.Host.ValueString()
	port := tfconfig.Port.ValueInt64()
	username := tfconfig.Username.ValueString()
	password := tfconfig.Password.ValueString()
	https := tfconfig.HTTPS.ValueBool()
	insecure := tfconfig.Insecure.ValueBool()
	tlsservername := tfconfig.TLSServerName.ValueString()
	cacert := tfconfig.CACert.ValueString()
	cakey := tfconfig.CAKey.ValueString()
	cert := tfconfig.Cert.ValueString()
	timeout := tfconfig.Timeout.ValueInt64()

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Hyper-V Host",
			"The provider cannot proceed as there is a missing or empty value for the Hyper-V host. "+
				"Set the host value in the configuration or use the HYPERV_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if port == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Missing Hyper-V Port",
			"The provider cannot proceed as there is a missing or empty value for the Hyper-V port. "+
				"Set the port value in the configuration or use the HYPERV_PORT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Hyper-V Username",
			"The provider cannot proceed as there is a missing or empty value for the Hyper-V username. "+
				"Set the username value in the configuration or use the HYPERV_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Hyper-V Password",
			"The provider cannot proceed as there is a missing or empty value for the Hyper-V password. "+
				"Set the password value in the configuration or use the HYPERV_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// endpoint := winrm.NewEndpoint(host, int(port), https, insecure, []byte(cacert), []byte(cert), []byte(cakey), time.Duration(timeout))
	// client, err := winrm.NewClient(endpoint, username, password)
	config := &hypervapi.Client{
		Host:          host,
		Port:          int(port),
		Username:      username,
		Password:      password,
		HTTPS:         https,
		Insecure:      insecure,
		TLSServerName: tlsservername,
		CACert:        []byte(cacert),
		CAKey:         []byte(cakey),
		Cert:          []byte(cert),
		Timeout:       time.Duration(timeout),
	}

	// Connect the client and handle any connection errors
	if err := config.Connect(); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Hyper-V Client",
			"An unexpected error occurred when creating the Hyper-V client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Hyper-V Client Error: "+err.Error(),
		)
		return
	}

	// Store *Client in ProviderData for access in resources and data sources
	resp.DataSourceData = config
	resp.ResourceData = config

}

func (p *HyperVProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVMResource,
	}
}

func (p *HyperVProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *HyperVProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HyperVProvider{
			version: version,
		}
	}
}
