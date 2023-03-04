package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/anitschke/photo-db-fs/types"
)

var (
	dbFactoryMu sync.RWMutex
	dbFactory   = make(map[string]DBConstructor)
)

type DBConstructor func(dbSource string) (DB, error)

// DB is an interface for interacting with a photo database to query information
// about photos within the database.
type DB interface {
	Photos(ctx context.Context, q types.Query) ([]types.Photo, error)
	RootTags(ctx context.Context) ([]types.Tag, error)
	ChildrenTags(ctx context.Context, parent types.Tag) ([]types.Tag, error)

	// Ratings should return a slice of ratings that will be used to render a
	// directory of folders based on these ratings. In most cases all possible
	// Ratings should be returned, if there is more than a "reasonable" number
	// of ratings exist then a subset of ratings that span the full range of
	// possible ratings should be returned. The slice of ratings should be in
	// acceding order.
	Ratings() []float64

	Close() error
}

// Register handles registration of a new DB type.
//
// We handle registration of new DB types similar to how SQL database drivers
// are registered. We expect that all DB types Register by calling Register with
// the init() function of their package. Then They should import their package
// into db/register. This way importing the db/register package into main will
// import all DB types
func Register(name string, dbCtor DBConstructor) {
	dbFactoryMu.Lock()
	defer dbFactoryMu.Unlock()
	if dbCtor == nil {
		panic("nil dbCtor")
	}
	if _, dup := dbFactory[name]; dup {
		panic("duplicate database type: " + name)
	}
	dbFactory[name] = dbCtor
}

func New(name string, dbSource string) (DB, error) {
	ctor, ok := dbFactory[name]
	if !ok {
		return nil, fmt.Errorf("dbCtor with name %q does not exist", name)
	}

	return ctor(dbSource)
}
