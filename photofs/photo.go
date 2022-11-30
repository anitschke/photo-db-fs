package photofs

import (
	"context"

	"github.com/anitschke/photo-db-fs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type photoNode struct {
	name string
	path string
}

var _ = (Node)((*photoNode)(nil))

func (n *photoNode) Name() string {
	return n.name
}

func (n *photoNode) Mode() uint32 {
	return fuse.S_IFLNK
}

func (n *photoNode) INode(ctx context.Context) (fs.InodeEmbedder, error) {
	symlink := &fs.MemSymlink{
		Data: []byte(n.path),
	}
	return symlink, nil
}

func photoSliceToNodeMap(photoSlice []types.Photo) (map[string]Node, error) {
	nodes := make([]Node, 0, len(photoSlice))
	for _, p := range photoSlice {
		nodes = append(nodes, &photoNode{
			name: p.UniqueStableName(),
			path: p.Path,
		})
	}
	ignoreDups := true
	return nodeSliceToNodeMap(nodes, ignoreDups)
}
