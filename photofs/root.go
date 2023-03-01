package photofs

import (
	"context"

	"github.com/anitschke/photo-db-fs/db"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/hanwen/go-fuse/v2/fs"
)

func NewRoot(ctx context.Context, db db.DB, queries []types.NamedQuery) (fs.InodeEmbedder, error) {
	return NewDirINode(ctx, &rootNode{db: db, queries: queries})
}

// rootNode is the root FUSE directory that contains the rest of our FUSE file
// system under it.
type rootNode struct {
	db      db.DB
	queries []types.NamedQuery
}

var _ = (DirNode)((*rootNode)(nil))

func (n *rootNode) Children(ctx context.Context) (map[string]Node, error) {
	nodes := []Node{
		&rootTagsNode{db: n.db},
		&rootQueriesNode{db: n.db, queries: n.queries},
	}
	ignoreDups := false
	return nodeSliceToNodeMap(nodes, ignoreDups)
}
