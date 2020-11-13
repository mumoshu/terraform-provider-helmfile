package helmfile

import (
	"errors"
	"golang.org/x/xerrors"
	"os"
	"path/filepath"
)

func shouldDiff(rs *ReleaseSet) (bool, error) {
	for _, path := range rs.SkipDiffOnMissingFiles {
		logf("processing %q in skip_diff_on_missing_files...", path)

		abs, err := filepath.Abs(path)
		if err != nil {
			return false, xerrors.Errorf("determining absolute path to %s: %w", path, err)
		}

		if _, err := os.Stat(abs); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				logf("detected missing file: %s", abs)

				return false, nil
			}

			return false, xerrors.Errorf("failed calling stat: %w", err)
		}

		logf("detected existing file: %s", abs)
	}

	return true, nil
}
