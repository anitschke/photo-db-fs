package testtools

import (
	"fmt"
	"os"
)

func MountPoint() (string, func(), error) {
	path, err := os.MkdirTemp("", "photo-db-fs_TEST_MOUNT_POINT")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		err := os.Remove(path)
		if err != nil {
			panic(fmt.Errorf("failed to cleanup %q: %w", path, err))
		}
	}
	return path, cleanup, nil
}
