package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RestAPIObjectDataSource{}
var _ datasource.DataSourceWithConfigure = &RestAPIObjectDataSource{}

// NewRestAPIObjectDataSource is a helper function to simplify the provider implementation.
func NewRestAPIObjectDataSource() datasource.DataSource {
	return &RestAPIObjectDataSource{}
}

// RestAPIObjectDataSource defines the data source implementation.
type RestAPIObjectDataSource struct {
	client *APIClient
}

func (d *RestAPIObjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (d *RestAPIObjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Performs a cURL get command on the specified url.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The API path on top of the base URL set in the provider that represents objects of this type on the API server.",
				Required:            true,
			},
			"search_path": schema.StringAttribute{
				MarkdownDescription: "The API path on top of the base URL set in the provider that represents the location to search for objects of this type on the API server. If not set, defaults to the value of path.",
				Optional:            true,
			},
			"query_string": schema.StringAttribute{
				MarkdownDescription: "An optional query string to send when performing the search.",
				Optional:            true,
			},
			"read_query_string": schema.StringAttribute{
				MarkdownDescription: "Defaults to `query_string` set on data source. This key allows setting a different or empty query string for reading the object.",
				Optional:            true,
			},
			"search_data": schema.StringAttribute{
				MarkdownDescription: "Valid JSON object to pass to search request as body.",
				Optional:            true,
			},
			"search_key": schema.StringAttribute{
				MarkdownDescription: "When reading search results from the API, this key is used to identify the specific record to read. This should be a unique record such as 'name'.",
				Required:            true,
			},
			"search_value": schema.StringAttribute{
				MarkdownDescription: "The value of 'search_key' will be compared to this value to determine if the correct object was found.",
				Required:            true,
			},
			"results_key": schema.StringAttribute{
				MarkdownDescription: "When issuing a GET to the path, this JSON key is used to locate the results array. The format is 'field/field/field'. Example: 'results/values'. If omitted, it is assumed the results coming back are already an array.",
				Optional:            true,
			},
			"id_attribute": schema.StringAttribute{
				MarkdownDescription: "Defaults to `id_attribute` set on the provider. Allows per-resource override of `id_attribute`.",
				Optional:            true,
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "Whether to emit verbose debug output while working with the API object on the server.",
				Optional:            true,
			},
			"api_data": schema.MapAttribute{
				MarkdownDescription: "After data from the API server is read, this map will include k/v pairs usable in other terraform resources as readable objects.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"api_response": schema.StringAttribute{
				MarkdownDescription: "The raw body of the HTTP response from the last read of the object.",
				Computed:            true,
			},
		},
	}
}

func (d *RestAPIObjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *RestAPIObjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RestAPIDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := data.Path.ValueString()
	searchPath := data.SearchPath.ValueString()
	queryString := data.QueryString.ValueString()
	debug := data.Debug.ValueBool()

	if debug {
		log.Printf("datasource_api_object.go: Data routine called.")
	}

	readQueryString := data.ReadQueryString.ValueString()
	if data.ReadQueryString.IsNull() || data.ReadQueryString.IsUnknown() {
		readQueryString = queryString
	}

	searchKey := data.SearchKey.ValueString()
	searchValue := data.SearchValue.ValueString()
	searchData := data.SearchData.ValueString()
	resultsKey := data.ResultsKey.ValueString()
	idAttribute := data.IDAttribute.ValueString()

	send := ""
	if len(searchData) > 0 {
		tmpData, _ := json.Marshal(searchData)
		send = string(tmpData)
		if debug {
			log.Printf("api_object.go: Using search data '%s'", send)
		}
	}

	if debug {
		log.Printf("datasource_api_object.go:\npath: %s\nsearch_path: %s\nquery_string: %s\nsearch_key: %s\nsearch_value: %s\nresults_key: %s\nid_attribute: %s", path, searchPath, queryString, searchKey, searchValue, resultsKey, idAttribute)
	}

	opts := &apiObjectOpts{
		path:        path,
		searchPath:  searchPath,
		debug:       debug,
		queryString: readQueryString,
		idAttribute: idAttribute,
	}

	obj, err := NewAPIObject(d.client, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API object", err.Error())
		return
	}

	if _, err := obj.findObject(queryString, searchKey, searchValue, resultsKey, send); err != nil {
		resp.Diagnostics.AddError("Failed to find object", err.Error())
		return
	}

	// Back to terraform-specific stuff. Create an api_object with the ID and refresh it object
	if debug {
		log.Printf("datasource_api_object.go: Attempting to construct api_object to refresh data")
	}

	err = obj.readObject()
	if err != nil {
		resp.Diagnostics.AddError("Failed to read object", err.Error())
		return
	}

	// Setting terraform ID tells terraform the object was created or it exists
	log.Printf("datasource_api_object.go: Data resource. Returned id is '%s'\n", obj.id)
	data.ID = types.StringValue(obj.id)

	// Set api_data from the object
	apiData := make(map[string]attr.Value)
	for k, v := range obj.apiData {
		apiData[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	data.APIData, _ = types.MapValue(types.StringType, apiData)
	data.APIResponse = types.StringValue(obj.apiResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
