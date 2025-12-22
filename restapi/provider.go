package restapi

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure RestAPIProvider satisfies various provider interfaces.
var _ provider.Provider = &RestAPIProvider{}

// RestAPIProvider defines the provider implementation.
type RestAPIProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RestAPIProvider{
			version: version,
		}
	}
}

func (p *RestAPIProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "restapi"
	resp.Version = p.version
}

func (p *RestAPIProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A provider for managing REST API resources.",
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				MarkdownDescription: "URI of the REST API endpoint. This serves as the base of all requests.",
				Required:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "When using https, this disables TLS verification of the host.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "When set, will use this username for BASIC auth to the API.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "When set, will use this password for BASIC auth to the API.",
				Optional:            true,
				Sensitive:           true,
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "A map of header names and values to set on all outbound requests.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"use_cookies": schema.BoolAttribute{
				MarkdownDescription: "Enable cookie jar to persist session.",
				Optional:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "When set, will cause requests taking longer than this time (in seconds) to be aborted.",
				Optional:            true,
			},
			"id_attribute": schema.StringAttribute{
				MarkdownDescription: "When set, this key will be used to operate on REST objects.",
				Optional:            true,
			},
			"create_method": schema.StringAttribute{
				MarkdownDescription: "Defaults to `POST`. The HTTP method used to CREATE objects of this type on the API server.",
				Optional:            true,
			},
			"read_method": schema.StringAttribute{
				MarkdownDescription: "Defaults to `GET`. The HTTP method used to READ objects of this type on the API server.",
				Optional:            true,
			},
			"update_method": schema.StringAttribute{
				MarkdownDescription: "Defaults to `PUT`. The HTTP method used to UPDATE objects of this type on the API server.",
				Optional:            true,
			},
			"destroy_method": schema.StringAttribute{
				MarkdownDescription: "Defaults to `DELETE`. The HTTP method used to DELETE objects of this type on the API server.",
				Optional:            true,
			},
			"copy_keys": schema.ListAttribute{
				MarkdownDescription: "When set, any PUT to the API for an object will copy these keys from the data the provider has gathered about the object.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"write_returns_object": schema.BoolAttribute{
				MarkdownDescription: "Set this when the API returns the object created on all write operations (POST, PUT).",
				Optional:            true,
			},
			"create_returns_object": schema.BoolAttribute{
				MarkdownDescription: "Set this when the API returns the object created only on creation operations (POST).",
				Optional:            true,
			},
			"xssi_prefix": schema.StringAttribute{
				MarkdownDescription: "Trim the xssi prefix from response string, if present, before parsing.",
				Optional:            true,
			},
			"rate_limit": schema.Float64Attribute{
				MarkdownDescription: "Set this to limit the number of requests per second made to the API.",
				Optional:            true,
			},
			"test_path": schema.StringAttribute{
				MarkdownDescription: "If set, the provider will issue a read_method request to this path after instantiation requiring a 200 OK response before proceeding.",
				Optional:            true,
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "Enabling this will cause lots of debug information to be printed to STDOUT by the API client.",
				Optional:            true,
			},
			"cert_string": schema.StringAttribute{
				MarkdownDescription: "When set with the key_string parameter, the provider will load a client certificate as a string for mTLS authentication.",
				Optional:            true,
			},
			"key_string": schema.StringAttribute{
				MarkdownDescription: "When set with the cert_string parameter, the provider will load a client certificate as a string for mTLS authentication.",
				Optional:            true,
				Sensitive:           true,
			},
			"cert_file": schema.StringAttribute{
				MarkdownDescription: "When set with the key_file parameter, the provider will load a client certificate as a file for mTLS authentication.",
				Optional:            true,
			},
			"key_file": schema.StringAttribute{
				MarkdownDescription: "When set with the cert_file parameter, the provider will load a client certificate as a file for mTLS authentication.",
				Optional:            true,
			},
			"root_ca_file": schema.StringAttribute{
				MarkdownDescription: "When set, the provider will load a root CA certificate as a file for mTLS authentication.",
				Optional:            true,
			},
			"root_ca_string": schema.StringAttribute{
				MarkdownDescription: "When set, the provider will load a root CA certificate as a string for mTLS authentication.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"oauth_client_credentials": schema.ListNestedBlock{
				MarkdownDescription: "Configuration for oauth client credential flow.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"oauth_client_id": schema.StringAttribute{
							MarkdownDescription: "OAuth client ID.",
							Required:            true,
						},
						"oauth_client_secret": schema.StringAttribute{
							MarkdownDescription: "OAuth client secret.",
							Required:            true,
							Sensitive:           true,
						},
						"oauth_token_endpoint": schema.StringAttribute{
							MarkdownDescription: "OAuth token endpoint URL.",
							Required:            true,
						},
						"oauth_scopes": schema.ListAttribute{
							MarkdownDescription: "OAuth scopes to request.",
							Optional:            true,
							ElementType:         types.StringType,
						},
						"endpoint_params": schema.MapAttribute{
							MarkdownDescription: "Additional key/values to pass to the underlying OAuth client library.",
							Optional:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (p *RestAPIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config RestAPIProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Handle environment variable defaults
	uri := config.URI.ValueString()
	if uri == "" {
		uri = os.Getenv("REST_API_URI")
	}
	if uri == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("uri"),
			"Missing REST API URI",
			"The provider cannot create the REST API client as there is a missing or empty value for the REST API URI. "+
				"Set the uri value in the configuration or use the REST_API_URI environment variable.",
		)
		return
	}

	// Build copy_keys slice
	copyKeys := make([]string, 0)
	if !config.CopyKeys.IsNull() && !config.CopyKeys.IsUnknown() {
		var copyKeysList []types.String
		config.CopyKeys.ElementsAs(ctx, &copyKeysList, false)
		for _, v := range copyKeysList {
			copyKeys = append(copyKeys, v.ValueString())
		}
	}

	// Build headers map
	headers := make(map[string]string)
	if !config.Headers.IsNull() && !config.Headers.IsUnknown() {
		var headersMap map[string]types.String
		config.Headers.ElementsAs(ctx, &headersMap, false)
		for k, v := range headersMap {
			headers[k] = v.ValueString()
		}
	}

	// Set rate limit default
	rateLimit := math.MaxFloat64
	if !config.RateLimit.IsNull() && !config.RateLimit.IsUnknown() {
		rateLimit = config.RateLimit.ValueFloat64()
	}

	opt := &apiClientOpt{
		uri:                 uri,
		insecure:            config.Insecure.ValueBool(),
		username:            config.Username.ValueString(),
		password:            config.Password.ValueString(),
		headers:             headers,
		useCookies:          config.UseCookies.ValueBool(),
		timeout:             int(config.Timeout.ValueInt64()),
		idAttribute:         config.IDAttribute.ValueString(),
		copyKeys:            copyKeys,
		writeReturnsObject:  config.WriteReturnsObject.ValueBool(),
		createReturnsObject: config.CreateReturnsObject.ValueBool(),
		xssiPrefix:          config.XSSIPrefix.ValueString(),
		rateLimit:           rateLimit,
		debug:               config.Debug.ValueBool(),
	}

	// Handle optional HTTP methods
	if !config.CreateMethod.IsNull() && !config.CreateMethod.IsUnknown() {
		opt.createMethod = config.CreateMethod.ValueString()
	}
	if !config.ReadMethod.IsNull() && !config.ReadMethod.IsUnknown() {
		opt.readMethod = config.ReadMethod.ValueString()
	}
	if !config.UpdateMethod.IsNull() && !config.UpdateMethod.IsUnknown() {
		opt.updateMethod = config.UpdateMethod.ValueString()
	}
	if !config.DestroyMethod.IsNull() && !config.DestroyMethod.IsUnknown() {
		opt.destroyMethod = config.DestroyMethod.ValueString()
	}

	// Handle OAuth configuration
	if !config.OAuthClientCredentials.IsNull() && !config.OAuthClientCredentials.IsUnknown() {
		var oauthConfigs []OAuthClientCredentialsModel
		config.OAuthClientCredentials.ElementsAs(ctx, &oauthConfigs, false)
		if len(oauthConfigs) > 0 {
			oauthConfig := oauthConfigs[0]
			opt.oauthClientID = oauthConfig.OAuthClientID.ValueString()
			opt.oauthClientSecret = oauthConfig.OAuthClientSecret.ValueString()
			opt.oauthTokenURL = oauthConfig.OAuthTokenEndpoint.ValueString()

			if !oauthConfig.OAuthScopes.IsNull() && !oauthConfig.OAuthScopes.IsUnknown() {
				var scopes []types.String
				oauthConfig.OAuthScopes.ElementsAs(ctx, &scopes, false)
				oauthScopesStr := make([]string, len(scopes))
				for i, s := range scopes {
					oauthScopesStr[i] = s.ValueString()
				}
				opt.oauthScopes = oauthScopesStr
			}

			if !oauthConfig.EndpointParams.IsNull() && !oauthConfig.EndpointParams.IsUnknown() {
				var paramsMap map[string]types.String
				oauthConfig.EndpointParams.ElementsAs(ctx, &paramsMap, false)
				setVals := url.Values{}
				for k, val := range paramsMap {
					setVals.Add(k, val.ValueString())
				}
				opt.oauthEndpointParams = setVals
			}
		}
	}

	// Handle TLS configuration
	if !config.CertFile.IsNull() && !config.CertFile.IsUnknown() {
		opt.certFile = config.CertFile.ValueString()
	}
	if !config.KeyFile.IsNull() && !config.KeyFile.IsUnknown() {
		opt.keyFile = config.KeyFile.ValueString()
	}
	if !config.CertString.IsNull() && !config.CertString.IsUnknown() {
		opt.certString = config.CertString.ValueString()
	}
	if !config.KeyString.IsNull() && !config.KeyString.IsUnknown() {
		opt.keyString = config.KeyString.ValueString()
	}
	if !config.RootCAFile.IsNull() && !config.RootCAFile.IsUnknown() {
		opt.rootCAFile = config.RootCAFile.ValueString()
	}
	if !config.RootCAString.IsNull() && !config.RootCAString.IsUnknown() {
		opt.rootCAString = config.RootCAString.ValueString()
	}

	client, err := NewAPIClient(opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create REST API Client",
			"An unexpected error occurred when creating the REST API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Test path validation
	if !config.TestPath.IsNull() && !config.TestPath.IsUnknown() {
		testPath := config.TestPath.ValueString()
		_, err := client.sendRequest(client.readMethod, testPath, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"REST API Test Path Failed",
				fmt.Sprintf("A test request to %v after setting up the provider did not return an OK response. "+
					"Is your configuration correct?\n\nError: %v", testPath, err),
			)
			return
		}
	}

	// Make the API client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *RestAPIProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRestAPIObjectResource,
	}
}

func (p *RestAPIProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRestAPIObjectDataSource,
	}
}
