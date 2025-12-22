package restapi

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RestAPIObjectModel describes the resource data model.
type RestAPIObjectModel struct {
	ID                     types.String `tfsdk:"id"`
	Path                   types.String `tfsdk:"path"`
	CreatePath             types.String `tfsdk:"create_path"`
	ReadPath               types.String `tfsdk:"read_path"`
	UpdatePath             types.String `tfsdk:"update_path"`
	DestroyPath            types.String `tfsdk:"destroy_path"`
	CreateMethod           types.String `tfsdk:"create_method"`
	ReadMethod             types.String `tfsdk:"read_method"`
	UpdateMethod           types.String `tfsdk:"update_method"`
	DestroyMethod          types.String `tfsdk:"destroy_method"`
	IDAttribute            types.String `tfsdk:"id_attribute"`
	ObjectID               types.String `tfsdk:"object_id"`
	Data                   types.String `tfsdk:"data"`
	Debug                  types.Bool   `tfsdk:"debug"`
	ReadSearch             types.Map    `tfsdk:"read_search"`
	QueryString            types.String `tfsdk:"query_string"`
	APIData                types.Map    `tfsdk:"api_data"`
	APIResponse            types.String `tfsdk:"api_response"`
	CreateResponse         types.String `tfsdk:"create_response"`
	ForceNew               types.List   `tfsdk:"force_new"`
	ReadData               types.String `tfsdk:"read_data"`
	UpdateData             types.String `tfsdk:"update_data"`
	DestroyData            types.String `tfsdk:"destroy_data"`
	IgnoreChangesTo        types.List   `tfsdk:"ignore_changes_to"`
	IgnoreAllServerChanges types.Bool   `tfsdk:"ignore_all_server_changes"`
}
