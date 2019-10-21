package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/sapcc/terraform-provider-ccloud/ccloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ccloud.Provider})
}
