package helmfile

import (
	"errors"
	"fmt"
	"github.com/mumoshu/shoal"
	"os"
	"path/filepath"
	"sync"
)

var shoalMu sync.Mutex

func prepareBinaries(fs *ReleaseSet) (*string, *string, error) {
	conf := shoal.Config{
		Git: shoal.Git{
			Provider: "go-git",
		},
	}

	rig := "https://github.com/fishworks/fish-food"

	helmBin := fs.HelmBin

	helmVersion := fs.HelmVersion

	installHelm := helmVersion != ""

	if installHelm {
		conf.Dependencies = append(conf.Dependencies,
			shoal.Dependency{
				Rig:     rig,
				Food:    "helm",
				Version: helmVersion,
			},
		)
		helmDiffVersion := fs.HelmDiffVersion
		if helmDiffVersion == "" {
			helmDiffVersion = "master"
		}
		conf.Helm.Plugins.Diff = helmDiffVersion
	}

	helmfileBin := fs.Bin

	helmfileVersion := fs.Version

	installHelmfile := helmfileVersion != ""

	if installHelmfile {
		conf.Dependencies = append(conf.Dependencies,
			shoal.Dependency{
				Rig:     rig,
				Food:    "helmfile",
				Version: helmfileVersion,
			},
		)
	}

	shoalMu.Lock()
	defer shoalMu.Unlock()

	s, err := shoal.New()
	if err != nil {
		return nil, nil, err
	}

	if len(conf.Dependencies) > 0 {
		if err := s.Init(); err != nil {
			return nil, nil, fmt.Errorf("initializing shoal: %w", err)
		}

		if err := s.InitGitProvider(conf); err != nil {
			return nil, nil, fmt.Errorf("initializing shoal git provider: %w", err)
		}

		wd, err := os.Getwd()
		if err != nil {
			return nil, nil, err
		}

		// TODO Any better place to do this?
		// This is for letting helm know about the location of helm plugins installed by shoal
		os.Setenv("XDG_DATA_HOME", filepath.Join(wd, ".shoal/Library"))

		if err := s.Sync(conf); err != nil {
			return nil, nil, err
		}
	}

	binPath := s.BinPath()

	if helmfileVersion != "" {
		helmfileBin = filepath.Join(binPath, "helmfile")
	}

	if helmVersion != "" {
		helmBin = filepath.Join(binPath, "helm")
	}

	if helmfileBin == "" {
		return nil, nil, errors.New("bug: helmfile_release_set.bin is missing")
	}

	return &helmfileBin, &helmBin, nil
}
