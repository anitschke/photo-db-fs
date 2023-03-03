package testtools

import (
	"fmt"
	"os"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
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

func MountTestFs(mountPoint string, root fs.InodeEmbedder) (*fuse.Server, error) {

	// To more closely replicate what we are doing in production so we can lock
	// down that minimal db calls are made when inode caching is turned on we
	// will also enable caching/timeout in tests. (but with an even longer value
	// to prevent any sort of sporadic)
	timeout := time.Hour

	return fs.Mount(mountPoint, root, &fs.Options{
		EntryTimeout:    &timeout,
		AttrTimeout:     &timeout,
		NegativeTimeout: &timeout,
	})
}
