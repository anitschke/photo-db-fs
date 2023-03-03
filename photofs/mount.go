//go:build linux

// photo-db-fs is built on top of FUSE using go-fuse for the implementation of
// this. FUSE is very much a linux/unix concept so I am doubtful this will work
// on Windows. At the moment I have only tested it out on linux although it will
// probably work on other *nix platforms and will probably work on OSX too as
// go-fuse claims limited support. If you want to try to port this to another
// platform remove the build tag at the top of this file and give it a try! Just
// be warned that your mileage may very.

package photofs

import (
	"context"
	"time"

	"github.com/anitschke/photo-db-fs/db"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Mount(ctx context.Context, mountPoint string, db db.DB, queries []types.NamedQuery) (*fuse.Server, error) {

	root, err := NewRoot(ctx, db, queries)
	if err != nil {
		zap.L().Fatal("failed to create file system", zap.Error(err))
	}

	fuseLogger, err := zap.NewStdLogAt(zap.L(), zapcore.DebugLevel)
	if err != nil {
		zap.L().Fatal("failed to create fuse logger", zap.Error(err))
	}

	// This timeout value controls the timeout for inode cache information to
	// hang around in the kernel. Without providing this setting the the mount
	// options it uses no caching in the kernel resulting in needing to do a db
	// look up for what seems like every time we go from a path to a inode. This
	// doesn't seem to be much of a problem in some situations like getting
	// directory contents from the terminal because that seems to hold on to
	// the inode file handle. But some programs like eog (gnome image viewer)
	// behave very poorly and take 30s or so just to open a photo because it
	// results in many many db lookups.
	//
	// Turning on caching seems to totally fix the issue I am seeing with eog,
	// so I am going to set this value to 10 min, I didn't do any special
	// experiments to get to this value, it just feels about right.
	//
	// TODO long term I need to look into if there is a way we can watch the
	// database file for changes and if there are invalidate the cache. go-fuse
	// seems to have a way to do this with an individual files viac
	// Inode.NotifyEntry, but I don't see a good way to ask the kernel for all
	// of inodes it has cached so we can invalidate all of them.
	timeout := 10 * time.Minute

	return fs.Mount(mountPoint, root, &fs.Options{
		EntryTimeout:    &timeout,
		AttrTimeout:     &timeout,
		NegativeTimeout: &timeout,

		Logger: fuseLogger,

		MountOptions: fuse.MountOptions{
			Name: "photo-db-fs",
		},
	})
}
