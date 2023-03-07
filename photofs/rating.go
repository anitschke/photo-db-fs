package photofs

import (
	"context"
	"strconv"

	"github.com/anitschke/photo-db-fs/db"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type ratingsParentNode struct {
	baseSelector types.Selector
	db           db.DB
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

	if len(ratings) == 0 {
		return map[string]Node{}, nil
	}

	children := make([]Node, 0, 2*len(ratings)-1)

	maxRating := ratings[len(ratings)-1]
	for _, r := range ratings {
		children = append(children, &ratingNode{baseSelector: n.baseSelector, operator: types.Equal, rating: r, db: n.db})
		if r != maxRating {
			children = append(children, &ratingNode{baseSelector: n.baseSelector, operator: types.GreaterThanOrEqual, rating: r, db: n.db})
		}
	}

	ignoreDups := false
	return nodeSliceToNodeMap(children, ignoreDups)
}

type ratingNode struct {
	baseSelector types.Selector
	operator     types.RelationalOperator
	rating       float64
	db           db.DB
}

var _ = (Node)((*ratingNode)(nil))
var _ = (DirNode)((*ratingNode)(nil))

func (n *ratingNode) Name() string {
	return string(n.operator) + strconv.Itoa(int(n.rating))
}

func (n *ratingNode) Mode() uint32 {
	return fuse.S_IFDIR
}

func (n *ratingNode) INode(ctx context.Context) (fs.InodeEmbedder, error) {
	return NewDirINode(ctx, n)
}

func (n *ratingNode) Children(ctx context.Context) (map[string]Node, error) {

	hasRatingSelector := types.HasRating{
		Operator: n.operator,
		Rating:   n.rating,
	}

	var selector types.Selector
	if n.baseSelector != nil {
		selector = types.And{
			Operands: []types.Selector{
				n.baseSelector,
				hasRatingSelector,
			},
		}
	} else {
		selector = hasRatingSelector
	}

	query := types.Query{
		Selector: selector,
	}

	childrenNodes := []Node{
		&queryNode{db: n.db, name: "photos", query: query},
	}
	ignoreDups := false
	return nodeSliceToNodeMap(childrenNodes, ignoreDups)
}
