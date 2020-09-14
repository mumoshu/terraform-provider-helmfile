package profile

import (
	"fmt"
	"github.com/pkg/profile"
	"os"
)

func Start() interface{ Stop() } {
	var opts []func(*profile.Profile)

	switch p := os.Getenv("TF_HELMFILE_PROFILE"); p {
	case "mem":
		opts = append(opts, profile.MemProfile)
	case "cpu":
		opts = append(opts, profile.CPUProfile)
	case "":
		// Do nothing
		return noopProfiler{}
	default:
		panic(fmt.Sprintf("Unsupported TF_HELMFILE_PROFILE=%s: Supported values are %q and %q", p, "mem", "cpu"))
	}

	if p := os.Getenv("TF_HELMFILE_PROFILE_PATH"); p != "" {
		opts = append(opts, profile.ProfilePath(p))
	}

	profiler := profile.Start(opts...)

	return profiler
}

type noopProfiler struct{}

func (_ noopProfiler) Stop() {

}
