package photofs

import (
	"context"
	"strconv"

	"github.com/anitschke/photo-db-fs/db"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

//xxx test

type ratingsParentNodeInfo struct {
	baseSelector types.Selector
	db           db.DB
}

type ratingsParentNode struct {
	ratingsParentNodeInfo
}

var _ = (Node)((*ratingsParentNode)(nil))
var _ = (DirNode)((*ratingsParentNode)(nil))

func (n *ratingsParentNode) Name() string {
	return "ratings"
}

func (n *ratingsParentNode) Mode() uint32 {
	return fuse.S_IFDIR
}

func (n *ratingsParentNode) INode(ctx context.Context) (fs.InodeEmbedder, error) {
	return NewDirINode(ctx, n)
}

func (n *ratingsParentNode) Children(ctx context.Context) (map[string]Node, error) {

	ratings := n.db.Ratings()
	children := make([]Node, 0, 2*len(ratings)-1)

	maxRating := ratings[len(ratings)-1]
	for _, r := range ratings {
		children = append(children, newRatingNode(types.Equal, r, n.baseSelector, n.db))
		if r != maxRating {
			children = append(children, newRatingNode(types.GreaterThanOrEqual, r, n.baseSelector, n.db))
		}
	}

	ignoreDups := false
	return nodeSliceToNodeMap(children, ignoreDups)
}

func newRatingNode(operator types.RelationalOperator, rating float64, baseSelector types.Selector, db db.DB) *queryNode {
	name := string(operator) + strconv.Itoa(int(rating))

	hasRatingSelector := types.HasRating{
		Operator: operator,
		Rating:   rating,
	}

	var selector types.Selector
	if baseSelector != nil {
		selector = types.And{
			Operands: []types.Selector{
				baseSelector,
				hasRatingSelector,
			},
		}
	} else {
		selector = hasRatingSelector
	}

	query := types.Query{
		Selector: selector,
	}

	return &queryNode{queryNodeInfo{db: db, name: name, query: query}}
}
