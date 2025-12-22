package main

import (
	"context"
	"log"

	"github.com/Mastercard/terraform-provider-restapi/restapi"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version will be set at build time
var version = "dev"

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/Mastercard/restapi",
	}

	err := providerserver.Serve(context.Background(), restapi.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
