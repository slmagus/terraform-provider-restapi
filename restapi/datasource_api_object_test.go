package restapi

import (
	"os"
	"testing"

	"github.com/Mastercard/terraform-provider-restapi/fakeserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRestApiDataSource_Basic(t *testing.T) {
	debug := false
	apiServerObjects := make(map[string]map[string]interface{})

	// Pre-populate fakeserver with an object to search for
	apiServerObjects["/api/objects"] = map[string]interface{}{
		"1234": map[string]interface{}{
			"id":    "1234",
			"first": "Foo",
			"last":  "Bar",
		},
	}

	svr := fakeserver.NewFakeServer(8083, apiServerObjects, true, debug, "")
	os.Setenv("REST_API_URI", "http://127.0.0.1:8083")
	svr.StartInBackground()
	defer svr.Shutdown()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "restapi" {
  uri = "http://127.0.0.1:8083/"
}

data "restapi_object" "test" {
  path         = "/api/objects"
  search_key   = "id"
  search_value = "1234"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.restapi_object.test", "id", "1234"),
				),
			},
		},
	})
}
