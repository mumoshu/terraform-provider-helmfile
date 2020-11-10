package helmfile

import (
	"errors"
	"golang.org/x/xerrors"
	"os"
)

func shouldDiff(rs *ReleaseSet) (bool, error) {
	for _, path := range rs.SkipDiffOnMissingFiles {
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				logf("detected missing file: %s", path)

				return false, nil
			}

			return false, xerrors.Errorf("failed calling stat: %w", err)
		}
	}

	return true, nil
}
