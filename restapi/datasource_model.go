package restapi

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RestAPIDataSourceModel describes the data source data model.
type RestAPIDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Path            types.String `tfsdk:"path"`
	SearchPath      types.String `tfsdk:"search_path"`
	QueryString     types.String `tfsdk:"query_string"`
	ReadQueryString types.String `tfsdk:"read_query_string"`
	SearchData      types.String `tfsdk:"search_data"`
	SearchKey       types.String `tfsdk:"search_key"`
	SearchValue     types.String `tfsdk:"search_value"`
	ResultsKey      types.String `tfsdk:"results_key"`
	IDAttribute     types.String `tfsdk:"id_attribute"`
	Debug           types.Bool   `tfsdk:"debug"`
	APIData         types.Map    `tfsdk:"api_data"`
	APIResponse     types.String `tfsdk:"api_response"`
}
