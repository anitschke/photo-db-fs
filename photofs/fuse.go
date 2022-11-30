package photofs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.uber.org/zap"
)

// go-fuse is great but it is still fairly low level in order to allow a lot
// more customization than we need. To make it a little easier for us to deal
// with these are some wrappers around go-fuse.

// dbERROR is the error that we will return when there is some sort of error
// reading the database or processing the data returned by the database. I when
// back and forth on what the correct error would be to return here, I think
// either EAGAIN or EIO would be correct but not sure. So I went with EIO since
// it seemed a little more generic?
const dbERROR = syscall.EIO

// Node is our interface for a file/directory within our virtual file system.
// Note that Node is NOT an INode but can rather get an INode, this is to make
// it so we can represent children of a directory without needing to actually
// create their INodes yet.
type Node interface {
	Name() string
	Mode() uint32
	INode(context.Context) (fs.InodeEmbedder, error)
}

// DirNode is our interface for a directory.
//
// Note that DirNode is not a Node because we only need to be a node for
// children of DirNode, but at the very top level our root node is not a child
// of any other DirNode but rather just mounted directly into the existing file
// system.
type DirNode interface {
	Children(context.Context) (map[string]Node, error)
}

type DirINode struct {
	fs.Inode
	children map[string]Node
}

var _ = (fs.NodeReaddirer)((*DirINode)(nil))

func NewDirINode(ctx context.Context, n DirNode) (fs.InodeEmbedder, error) {
	// We will ask for the children of this dir node once when we create the
	// INode and then cache the results. This allows us to minimize DB lookups.
	// We also do it here instead of an on demand caching because it allows us to
	// avoid needing a mutex in an on demand function since concurrent calls to
	// Lookup are possible.

	c, err := n.Children(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup directory children: %w", err)
	}
	return &DirINode{
		children: c,
	}, nil
}

func (n *DirINode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r := make([]fuse.DirEntry, 0, len(n.children))
	for _, c := range n.children {
		r = append(r, fuse.DirEntry{
			Name: c.Name(),
			Mode: c.Mode(),
		})
	}
	return fs.NewListDirStream(r), 0
}

var _ = (fs.NodeLookuper)((*DirINode)(nil))

func (n *DirINode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	c, ok := n.children[name]
	if !ok {
		return nil, syscall.ENOENT
	}

	stable := fs.StableAttr{
		Mode: c.Mode(),
	}

	operations, err := c.INode(ctx)
	if err != nil {
		zap.L().Error("error getting child INode", zap.Any("child", c), zap.Error(err))
		return nil, dbERROR
	}

	childNode := n.NewInode(ctx, operations, stable)
	return childNode, 0
}

func nodeSliceToNodeMap(nodeSlice []Node, ignoreDups bool) (map[string]Node, error) {
	nodeMap := make(map[string]Node, len(nodeSlice))

	for _, n := range nodeSlice {
		name := n.Name()
		if existing, exist := nodeMap[name]; exist {

			// There is the potential for name collisions. In most cases we will
			// error out in these cases since they should never happen, but for
			// some cases it might happen with photos and we will just ignore
			// the dups. see discussion in types.Photo.ID. In these cases the
			// cause (at least for digikam) is both photos have the same
			// content. So we can just safely ignore one of the photos. We will
			// still issue a warning, but it shouldn't be that big a deal for
			// most users.
			if ignoreDups {
				zap.L().Warn("detected node with duplicate name", zap.Any("existing", existing), zap.Any("new", n))
			} else {
				return nil, fmt.Errorf("detected node with duplicate name %q", name)
			}
		} else {
			nodeMap[name] = n
		}
	}
	return nodeMap, nil
}
