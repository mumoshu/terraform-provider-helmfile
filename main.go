package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/mumoshu/terraform-provider-helmfile/pkg/helmfile"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: helmfile.Provider})
}
