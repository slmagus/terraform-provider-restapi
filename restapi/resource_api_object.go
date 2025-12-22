package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RestAPIObjectResource{}
var _ resource.ResourceWithImportState = &RestAPIObjectResource{}
var _ resource.ResourceWithConfigure = &RestAPIObjectResource{}

// NewRestAPIObjectResource is a helper function to simplify the provider implementation.
func NewRestAPIObjectResource() resource.Resource {
	return &RestAPIObjectResource{}
}

// RestAPIObjectResource defines the resource implementation.
type RestAPIObjectResource struct {
	client *APIClient
}

func (r *RestAPIObjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (r *RestAPIObjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Consider data sensitive if env variable is set to true.
	isDataSensitive, _ := strconv.ParseBool(GetEnvOrDefault("API_DATA_IS_SENSITIVE", "false"))

	resp.Schema = schema.Schema{
		MarkdownDescription: "Acting as a wrapper of cURL, this object supports POST, GET, PUT and DELETE on the specified url",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The API path on top of the base URL set in the provider that represents objects of this type on the API server.",
				Required:            true,
			},
			"create_path": schema.StringAttribute{
				MarkdownDescription: "Defaults to `path`. The API path that represents where to CREATE (POST) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object if the data contains the `id_attribute`.",
				Optional:            true,
			},
			"read_path": schema.StringAttribute{
				MarkdownDescription: "Defaults to `path/{id}`. The API path that represents where to READ (GET) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:            true,
			},
			"update_path": schema.StringAttribute{
				MarkdownDescription: "Defaults to `path/{id}`. The API path that represents where to UPDATE (PUT) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:            true,
			},
			"destroy_path": schema.StringAttribute{
				MarkdownDescription: "Defaults to `path/{id}`. The API path that represents where to DESTROY (DELETE) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:            true,
			},
			"create_method": schema.StringAttribute{
				MarkdownDescription: "Defaults to `create_method` set on the provider. Allows per-resource override of `create_method`.",
				Optional:            true,
			},
			"read_method": schema.StringAttribute{
				MarkdownDescription: "Defaults to `read_method` set on the provider. Allows per-resource override of `read_method`.",
				Optional:            true,
			},
			"update_method": schema.StringAttribute{
				MarkdownDescription: "Defaults to `update_method` set on the provider. Allows per-resource override of `update_method`.",
				Optional:            true,
			},
			"destroy_method": schema.StringAttribute{
				MarkdownDescription: "Defaults to `destroy_method` set on the provider. Allows per-resource override of `destroy_method`.",
				Optional:            true,
			},
			"id_attribute": schema.StringAttribute{
				MarkdownDescription: "Defaults to `id_attribute` set on the provider. Allows per-resource override of `id_attribute`.",
				Optional:            true,
			},
			"object_id": schema.StringAttribute{
				MarkdownDescription: "Defaults to the id learned by the provider during normal operations and `id_attribute`. Allows you to set the id manually.",
				Optional:            true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "Valid JSON object that this provider will manage with the API server.",
				Required:            true,
				Sensitive:           isDataSensitive,
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "Whether to emit verbose debug output while working with the API object on the server.",
				Optional:            true,
			},
			"read_search": schema.MapAttribute{
				MarkdownDescription: "Custom search for `read_path`. This map will take `search_data`, `search_key`, `search_value`, `results_key` and `query_string`.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"query_string": schema.StringAttribute{
				MarkdownDescription: "Query string to be included in the path.",
				Optional:            true,
			},
			"api_data": schema.MapAttribute{
				MarkdownDescription: "After data from the API server is read, this map will include k/v pairs usable in other terraform resources as readable objects.",
				Computed:            true,
				Sensitive:           isDataSensitive,
				ElementType:         types.StringType,
			},
			"api_response": schema.StringAttribute{
				MarkdownDescription: "The raw body of the HTTP response from the last read of the object.",
				Computed:            true,
				Sensitive:           isDataSensitive,
			},
			"create_response": schema.StringAttribute{
				MarkdownDescription: "The raw body of the HTTP response returned when creating the object.",
				Computed:            true,
				Sensitive:           isDataSensitive,
			},
			"force_new": schema.ListAttribute{
				MarkdownDescription: "Any changes to these values will result in recreating the resource instead of updating.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"read_data": schema.StringAttribute{
				MarkdownDescription: "Valid JSON object to pass during read requests.",
				Optional:            true,
				Sensitive:           isDataSensitive,
			},
			"update_data": schema.StringAttribute{
				MarkdownDescription: "Valid JSON object to pass during update requests.",
				Optional:            true,
				Sensitive:           isDataSensitive,
			},
			"destroy_data": schema.StringAttribute{
				MarkdownDescription: "Valid JSON object to pass during destroy requests.",
				Optional:            true,
				Sensitive:           isDataSensitive,
			},
			"ignore_changes_to": schema.ListAttribute{
				MarkdownDescription: "A list of fields to which remote changes will be ignored.",
				Optional:            true,
				Sensitive:           isDataSensitive,
				ElementType:         types.StringType,
			},
			"ignore_all_server_changes": schema.BoolAttribute{
				MarkdownDescription: "By default Terraform will attempt to revert changes to remote resources. Set this to 'true' to ignore any remote changes. Default: false",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *RestAPIObjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *RestAPIObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RestAPIObjectModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts, err := r.buildAPIObjectOpts(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build API object options", err.Error())
		return
	}

	debug := data.Debug.ValueBool()
	obj, err := NewAPIObject(r.client, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API object", err.Error())
		return
	}

	if debug {
		log.Printf("resource_api_object.go: Create routine called. Object built:\n%s\n", obj.toString())
	}

	err = obj.createObject()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create object via API", err.Error())
		return
	}

	// Update state from API response
	data.ID = types.StringValue(obj.id)
	r.setResourceState(ctx, obj, &data)
	data.CreateResponse = types.StringValue(obj.apiResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestAPIObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RestAPIObjectModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts, err := r.buildAPIObjectOpts(ctx, &data)
	if err != nil {
		if strings.Contains(err.Error(), "error parsing data provided") {
			log.Printf("resource_api_object.go: WARNING! The data passed from Terraform's state is invalid! %v", err)
			log.Printf("resource_api_object.go: Continuing with partially constructed object...")
		} else {
			resp.Diagnostics.AddError("Failed to build API object options", err.Error())
			return
		}
	}

	debug := data.Debug.ValueBool()
	obj, err := NewAPIObject(r.client, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API object", err.Error())
		return
	}

	if debug {
		log.Printf("resource_api_object.go: Read routine called. Object built:\n%s\n", obj.toString())
	}

	err = obj.readObject()
	if err != nil {
		// Object doesn't exist, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	log.Printf("resource_api_object.go: Read resource. Returned id is '%s'\n", obj.id)
	data.ID = types.StringValue(obj.id)
	r.setResourceState(ctx, obj, &data)

	// Check whether the remote resource has changed
	if !data.IgnoreAllServerChanges.ValueBool() {
		ignoreList := []string{}
		if !data.IgnoreChangesTo.IsNull() && !data.IgnoreChangesTo.IsUnknown() {
			var ignoreItems []types.String
			data.IgnoreChangesTo.ElementsAs(ctx, &ignoreItems, false)
			for _, s := range ignoreItems {
				ignoreList = append(ignoreList, s.ValueString())
			}
		}

		modifiedResource, hasDifferences := getDelta(obj.data, obj.apiData, ignoreList)
		if hasDifferences {
			log.Printf("resource_api_object.go: Found differences in remote resource\n")
			encoded, err := json.Marshal(modifiedResource)
			if err != nil {
				resp.Diagnostics.AddError("Failed to encode modified resource", err.Error())
				return
			}
			data.Data = types.StringValue(string(encoded))
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestAPIObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RestAPIObjectModel
	var state RestAPIObjectModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve create_response from state
	data.CreateResponse = state.CreateResponse

	opts, err := r.buildAPIObjectOpts(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build API object options", err.Error())
		return
	}

	debug := data.Debug.ValueBool()
	obj, err := NewAPIObject(r.client, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API object", err.Error())
		return
	}

	// If copy_keys is not empty, we have to grab the latest data
	if len(r.client.copyKeys) > 0 {
		err = obj.readObject()
		if err != nil {
			resp.Diagnostics.AddError("Failed to read object before update", err.Error())
			return
		}
	}

	if debug {
		log.Printf("resource_api_object.go: Update routine called. Object built:\n%s\n", obj.toString())
	}

	err = obj.updateObject()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update object via API", err.Error())
		return
	}

	r.setResourceState(ctx, obj, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RestAPIObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RestAPIObjectModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts, err := r.buildAPIObjectOpts(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build API object options", err.Error())
		return
	}

	debug := data.Debug.ValueBool()
	obj, err := NewAPIObject(r.client, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API object", err.Error())
		return
	}

	if debug {
		log.Printf("resource_api_object.go: Delete routine called. Object built:\n%s\n", obj.toString())
	}

	err = obj.deleteObject()
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			// 404 means it doesn't exist. Call that good enough
			return
		}
		resp.Diagnostics.AddError("Failed to delete object via API", err.Error())
		return
	}
}

func (r *RestAPIObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	input := req.ID

	hasTrailingSlash := strings.HasSuffix(input, "/")
	var n int
	if hasTrailingSlash {
		n = strings.LastIndex(input[0:len(input)-1], "/")
	} else {
		n = strings.LastIndex(input, "/")
	}

	if n == -1 {
		resp.Diagnostics.AddError(
			"Invalid Import Path",
			fmt.Sprintf("Invalid path to import api_object '%s' - must be /<full path from server root>/<object id>", input),
		)
		return
	}

	apiPath := input[0:n]

	var id string
	if hasTrailingSlash {
		id = input[n+1 : len(input)-1]
	} else {
		id = input[n+1:]
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, fwpath.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, fwpath.Root("path"), apiPath)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, fwpath.Root("data"), fmt.Sprintf(`{ "id": "%s" }`, id))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, fwpath.Root("debug"), true)...)
}

// Helper function to build API object options from model
func (r *RestAPIObjectResource) buildAPIObjectOpts(ctx context.Context, data *RestAPIObjectModel) (*apiObjectOpts, error) {
	opts := &apiObjectOpts{
		path: data.Path.ValueString(),
	}

	if !data.IDAttribute.IsNull() && !data.IDAttribute.IsUnknown() {
		opts.idAttribute = data.IDAttribute.ValueString()
	}

	if !data.ObjectID.IsNull() && !data.ObjectID.IsUnknown() {
		opts.id = data.ObjectID.ValueString()
	} else {
		opts.id = data.ID.ValueString()
	}

	log.Printf("resource_rest_api.go: buildAPIObjectOpts routine called for id '%s'\n", opts.id)

	if !data.CreatePath.IsNull() && !data.CreatePath.IsUnknown() {
		opts.postPath = data.CreatePath.ValueString()
	}
	if !data.ReadPath.IsNull() && !data.ReadPath.IsUnknown() {
		opts.getPath = data.ReadPath.ValueString()
	}
	if !data.UpdatePath.IsNull() && !data.UpdatePath.IsUnknown() {
		opts.putPath = data.UpdatePath.ValueString()
	}
	if !data.CreateMethod.IsNull() && !data.CreateMethod.IsUnknown() {
		opts.createMethod = data.CreateMethod.ValueString()
	}
	if !data.ReadMethod.IsNull() && !data.ReadMethod.IsUnknown() {
		opts.readMethod = data.ReadMethod.ValueString()
	}
	if !data.ReadData.IsNull() && !data.ReadData.IsUnknown() {
		opts.readData = data.ReadData.ValueString()
	}
	if !data.UpdateMethod.IsNull() && !data.UpdateMethod.IsUnknown() {
		opts.updateMethod = data.UpdateMethod.ValueString()
	}
	if !data.UpdateData.IsNull() && !data.UpdateData.IsUnknown() {
		opts.updateData = data.UpdateData.ValueString()
	}
	if !data.DestroyMethod.IsNull() && !data.DestroyMethod.IsUnknown() {
		opts.destroyMethod = data.DestroyMethod.ValueString()
	}
	if !data.DestroyData.IsNull() && !data.DestroyData.IsUnknown() {
		opts.destroyData = data.DestroyData.ValueString()
	}
	if !data.DestroyPath.IsNull() && !data.DestroyPath.IsUnknown() {
		opts.deletePath = data.DestroyPath.ValueString()
	}
	if !data.QueryString.IsNull() && !data.QueryString.IsUnknown() {
		opts.queryString = data.QueryString.ValueString()
	}

	// Handle read_search map
	if !data.ReadSearch.IsNull() && !data.ReadSearch.IsUnknown() {
		var readSearchMap map[string]types.String
		data.ReadSearch.ElementsAs(ctx, &readSearchMap, false)
		readSearch := make(map[string]string)
		for k, v := range readSearchMap {
			readSearch[k] = v.ValueString()
		}
		opts.readSearch = readSearch
	}

	opts.data = data.Data.ValueString()
	opts.debug = data.Debug.ValueBool()

	return opts, nil
}

// Helper function to set resource state from API object
func (r *RestAPIObjectResource) setResourceState(ctx context.Context, obj *APIObject, data *RestAPIObjectModel) {
	apiData := make(map[string]attr.Value)
	for k, v := range obj.apiData {
		apiData[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	data.APIData, _ = types.MapValue(types.StringType, apiData)
	data.APIResponse = types.StringValue(obj.apiResponse)
}
