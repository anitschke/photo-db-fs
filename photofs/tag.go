package photofs

import (
	"context"
	"fmt"
	"path"

	"github.com/anitschke/photo-db-fs/db"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// rootTagsNode is the top FUSE directory that contains the whole tag hierarchy
// under it.
type rootTagsNode struct {
	db db.DB
}

var _ = (Node)((*rootTagsNode)(nil))
var _ = (DirNode)((*rootTagsNode)(nil))

func (n *rootTagsNode) Name() string {
	return "tags"
}

func (n *rootTagsNode) Mode() uint32 {
	return fuse.S_IFDIR
}

func (n *rootTagsNode) INode(ctx context.Context) (fs.InodeEmbedder, error) {
	return NewDirINode(ctx, n)
}

func (n *rootTagsNode) Children(ctx context.Context) (map[string]Node, error) {
	rootTags, err := n.db.RootTags(ctx)
	if err != nil {
		return nil, err
	}
	return tagSliceToNodeMap(n.db, rootTags)
}

type tagNodeInfo struct {
	tag types.Tag
	db  db.DB
}

type tagNode struct {
	tagNodeInfo
}

var _ = (Node)((*tagNode)(nil))
var _ = (DirNode)((*tagNode)(nil))

func (n *tagNode) Name() string {
	return n.tag.Name()
}

func (n *tagNode) Mode() uint32 {
	return fuse.S_IFDIR
}

func (n *tagNode) INode(ctx context.Context) (fs.InodeEmbedder, error) {
	return NewDirINode(ctx, n)
}

func (n *tagNode) Children(ctx context.Context) (map[string]Node, error) {
	childrenNodes := []Node{
		&childTagsNode{tagNodeInfo: n.tagNodeInfo},
		&queryNode{queryNodeInfo{db: n.db, name: "photos", query: types.Query{Selector: types.HasTag{Tag: n.tag}}}},
	}
	ignoreDups := false
	return nodeSliceToNodeMap(childrenNodes, ignoreDups)
}

type childTagsNode struct {
	tagNodeInfo
}

var _ = (Node)((*childTagsNode)(nil))
var _ = (DirNode)((*childTagsNode)(nil))

func (n *childTagsNode) Name() string {
	return "tags"
}

func (n *childTagsNode) Mode() uint32 {
	return fuse.S_IFDIR
}

func (n *childTagsNode) INode(ctx context.Context) (fs.InodeEmbedder, error) {
	return NewDirINode(ctx, n)
}

func (n *childTagsNode) Children(ctx context.Context) (map[string]Node, error) {

	children, err := n.db.ChildrenTags(ctx, n.tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags that are children of tag %q: %w", path.Join(n.tag.Path...), err)
	}
	return tagSliceToNodeMap(n.db, children)
}

func tagSliceToNodeMap(db db.DB, tagSlice []types.Tag) (map[string]Node, error) {
	nodes := make([]Node, 0, len(tagSlice))
	for _, t := range tagSlice {
		nodes = append(nodes, &tagNode{tagNodeInfo: tagNodeInfo{db: db, tag: t}})
	}
	ignoreDups := false
	return nodeSliceToNodeMap(nodes, ignoreDups)
}
