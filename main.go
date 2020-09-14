package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/mumoshu/terraform-provider-helmfile/pkg/helmfile"
	"github.com/mumoshu/terraform-provider-helmfile/pkg/profile"
)

func main() {
	defer profile.Start().Stop()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: helmfile.Provider})
}
