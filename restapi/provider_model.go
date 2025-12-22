package restapi

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RestAPIProviderModel describes the provider data model.
type RestAPIProviderModel struct {
	URI                    types.String  `tfsdk:"uri"`
	Insecure               types.Bool    `tfsdk:"insecure"`
	Username               types.String  `tfsdk:"username"`
	Password               types.String  `tfsdk:"password"`
	Headers                types.Map     `tfsdk:"headers"`
	UseCookies             types.Bool    `tfsdk:"use_cookies"`
	Timeout                types.Int64   `tfsdk:"timeout"`
	IDAttribute            types.String  `tfsdk:"id_attribute"`
	CreateMethod           types.String  `tfsdk:"create_method"`
	ReadMethod             types.String  `tfsdk:"read_method"`
	UpdateMethod           types.String  `tfsdk:"update_method"`
	DestroyMethod          types.String  `tfsdk:"destroy_method"`
	CopyKeys               types.List    `tfsdk:"copy_keys"`
	WriteReturnsObject     types.Bool    `tfsdk:"write_returns_object"`
	CreateReturnsObject    types.Bool    `tfsdk:"create_returns_object"`
	XSSIPrefix             types.String  `tfsdk:"xssi_prefix"`
	RateLimit              types.Float64 `tfsdk:"rate_limit"`
	TestPath               types.String  `tfsdk:"test_path"`
	Debug                  types.Bool    `tfsdk:"debug"`
	OAuthClientCredentials types.List    `tfsdk:"oauth_client_credentials"`
	CertString             types.String  `tfsdk:"cert_string"`
	KeyString              types.String  `tfsdk:"key_string"`
	CertFile               types.String  `tfsdk:"cert_file"`
	KeyFile                types.String  `tfsdk:"key_file"`
	RootCAFile             types.String  `tfsdk:"root_ca_file"`
	RootCAString           types.String  `tfsdk:"root_ca_string"`
}

// OAuthClientCredentialsModel describes the nested OAuth configuration.
type OAuthClientCredentialsModel struct {
	OAuthClientID      types.String `tfsdk:"oauth_client_id"`
	OAuthClientSecret  types.String `tfsdk:"oauth_client_secret"`
	OAuthTokenEndpoint types.String `tfsdk:"oauth_token_endpoint"`
	OAuthScopes        types.List   `tfsdk:"oauth_scopes"`
	EndpointParams     types.Map    `tfsdk:"endpoint_params"`
}
