package restapi

import (
	"testing"

	"github.com/Mastercard/terraform-provider-restapi/fakeserver"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"restapi": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider(t *testing.T) {
	// Create and validate a provider instance
	p := New("test")()
	if p == nil {
		t.Fatalf("Provider returned nil")
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = New("test")()
}

func TestResourceProvider_RequireTestPath(t *testing.T) {
	debug := false
	apiServerObjects := make(map[string]map[string]interface{})

	svr := fakeserver.NewFakeServer(8085, apiServerObjects, true, debug, "")
	svr.StartInBackground()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "restapi" {
  uri = "http://127.0.0.1:8085/"
  test_path = "/api/objects"
}

data "restapi_object" "test" {
  path = "/api/objects"
  search_key = "id"
  search_value = "test"
}
`,
				ExpectError: nil,
			},
		},
	})

	svr.Shutdown()
}

func testAccPreCheck(t *testing.T) {
	// Add any pre-check logic here
}

func providerConfig() string {
	return `
provider "restapi" {
  uri = "http://127.0.0.1:8082/"
}
`
}
