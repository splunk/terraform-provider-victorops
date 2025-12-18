package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-victorops/victorops"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: victorops.Provider,
	})
}
