package main

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/mumoshu/terraform-provider-helmfile/pkg/helmfile"
	"github.com/pkg/profile"
	"os"
)

func main() {
	defer StartProfiling().Stop()

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: helmfile.Provider})
}

func StartProfiling() interface{ Stop() } {
	var opts []func(*profile.Profile)

	switch p := os.Getenv("TF_HELMFILE_PROFILE"); p {
	case "mem":
		opts = append(opts, profile.MemProfile)
	case "cpu":
		opts = append(opts, profile.CPUProfile)
	case "":
		// Do nothing
	default:
		panic(fmt.Sprintf("Unsupported TF_HELMFILE_PROFILE=%s: Supported values are %q and %q", p, "mem", "cpu"))
	}

	if p := os.Getenv("TF_HELMFILE_PROFILE_PATH"); p != "" {
		opts = append(opts, profile.ProfilePath(p))
	}

	profiler := profile.Start(opts...)

	return profiler
}
