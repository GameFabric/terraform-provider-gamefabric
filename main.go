package main

import (
	"context"
	"flag"
	"log"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version string

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address:         "registry.terraform.io/GameFabric/gamefabric",
		Debug:           debug,
		ProtocolVersion: 6,
	}
	if debug {
		opts.Debug = true
	}

	serveErr := providerserver.Serve(context.Background(), provider.New(version), opts)
	if serveErr != nil {
		log.Fatal(serveErr.Error())
	}
}
