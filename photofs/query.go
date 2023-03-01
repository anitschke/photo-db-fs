package photofs

import (
	"context"
	"fmt"

	"github.com/anitschke/photo-db-fs/db"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

//xxx test

// rootQueriesNode is the top FUSE directory that contains a folder of all the
// queries
type rootQueriesNode struct {
	db      db.DB
	queries []types.NamedQuery
}

var _ = (Node)((*rootQueriesNode)(nil))
var _ = (DirNode)((*rootQueriesNode)(nil))

func (n *rootQueriesNode) Name() string {
	return "queries"
}

func (n *rootQueriesNode) Mode() uint32 {
	return fuse.S_IFDIR
}

func (n *rootQueriesNode) INode(ctx context.Context) (fs.InodeEmbedder, error) {
	return NewDirINode(ctx, n)
}

func (n *rootQueriesNode) Children(ctx context.Context) (map[string]Node, error) {
	nodes := make([]Node, 0, len(n.queries))
	for _, q := range n.queries {
		nodes = append(nodes, &queryNode{queryNodeInfo: queryNodeInfo{db: n.db, query: q}})
	}
	ignoreDups := false
	return nodeSliceToNodeMap(nodes, ignoreDups)
}

type queryNodeInfo struct {
	query types.NamedQuery
	db    db.DB
}

type queryNode struct {
	queryNodeInfo
}

var _ = (Node)((*queryNode)(nil))
var _ = (DirNode)((*queryNode)(nil))

func (n *queryNode) Name() string {
	return n.query.Name
}

func (n *queryNode) Mode() uint32 {
	return fuse.S_IFDIR
}

func (n *queryNode) INode(ctx context.Context) (fs.InodeEmbedder, error) {
	return NewDirINode(ctx, n)
}

func (n *queryNode) Children(ctx context.Context) (map[string]Node, error) {
	children, err := n.db.Photos(ctx, n.query.Query)
	if err != nil {
		return nil, fmt.Errorf("failed perform named query %q: %w", n.query.Name, err)
	}
	return photoSliceToNodeMap(children)
}
