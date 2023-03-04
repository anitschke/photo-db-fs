package digikam

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/anitschke/photo-db-fs/db"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/anitschke/photo-db-fs/utils"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func init() {
	db.Register("digikam-sqlite", func(dbSource string) (db.DB, error) {
		return NewDigikamSqliteDatabase(dbSource)
	})
}

type DigikamSQLDatabase struct {
	db *sql.DB
}

var _ = (db.DB)((*DigikamSQLDatabase)(nil))

func NewDigikamSqliteDatabase(filePath string) (*DigikamSQLDatabase, error) {
	return NewDigikamSQLDatabase("sqlite3", filePath)
}

func NewDigikamSQLDatabase(driver string, filePath string) (*DigikamSQLDatabase, error) {
	connectionString := "file:" + filePath + "?mode=ro"
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}

	return &DigikamSQLDatabase{
		db: db,
	}, nil
}

func (db *DigikamSQLDatabase) Photos(ctx context.Context, q types.Query) ([]types.Photo, error) {
	zap.L().Debug("db query photos", zap.Any("query", q))

	queryString, parameters, err := buildDigikamPhotoQuery(q)
	if err != nil {
		return nil, err
	}

	queryString = addCountToQuery(queryString)

	zap.L().Debug("db query", zap.String("query", queryString), zap.Any("parameters", parameters))
	rows, err := db.db.QueryContext(ctx, queryString, parameters...)
	if err != nil {
		return nil, err
	}
	defer utils.CloseAndLogErrors(rows)

	// don't make until we know how big to make our slice (increasing capacity
	// of slices is expensive)
	var photos []types.Photo

	for rows.Next() {
		// root, path, name, uniqueHash

		var nPhotos int
		var root string
		var path string
		var name string
		var uniqueHash string
		err = rows.Scan(&nPhotos, &root, &path, &name, &uniqueHash)
		if err != nil {
			return nil, err
		}

		fullPath := filepath.Join(root, path, name)

		p := types.Photo{
			Path: fullPath,
			ID:   uniqueHash,
		}

		// If the tags slice doesn't exist yet then make it with enough elements
		// so we aren't constantly resizing on every append
		if photos == nil {
			photos = make([]types.Photo, 0, nPhotos)
		}

		photos = append(photos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	zap.L().Debug("db query photos passed", zap.Any("query", q), zap.Int("resultCount", len(photos)))
	return photos, nil
}

func (db *DigikamSQLDatabase) RootTags(ctx context.Context) ([]types.Tag, error) {
	zap.L().Debug("db query root tags")

	where := "pid=0"
	parentPath := []string{}
	parameters := make([]any, 0)
	return db.tags(ctx, parentPath, where, parameters)
}

func (db *DigikamSQLDatabase) ChildrenTags(ctx context.Context, p types.Tag) ([]types.Tag, error) {
	zap.L().Debug("db query children tags", zap.Any("parent", p))

	parentQuery, parameters, err := tagIDSubquery(p)
	if err != nil {
		return nil, err
	}

	where := "pid=" + parentQuery

	return db.tags(ctx, p.Path, where, parameters)
}

func (db *DigikamSQLDatabase) tags(ctx context.Context, parentPath []string, where string, parameters []any) ([]types.Tag, error) {
	q := "SELECT name FROM Tags WHERE " + where
	q = addCountToQuery(q)

	zap.L().Debug("db query", zap.String("query", q), zap.Any("parameters", parameters))
	rows, err := db.db.QueryContext(ctx, q, parameters...)
	if err != nil {
		return nil, err
	}
	defer utils.CloseAndLogErrors(rows)

	// don't make until we know how big to make our slice (increasing capacity
	// of slices is expensive)
	var tags []types.Tag

	for rows.Next() {
		var nTags int
		var name string
		err = rows.Scan(&nTags, &name)
		if err != nil {
			return nil, err
		}

		t := types.Tag{}
		t.Path = make([]string, len(parentPath), len(parentPath)+1)
		copy(t.Path, parentPath)
		t.Path = append(t.Path, name)

		// If the tags slice doesn't exist yet then make it with enough elements
		// so we aren't constantly resizing on every append
		if tags == nil {
			tags = make([]types.Tag, 0, nTags)
		}

		tags = append(tags, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	zap.L().Debug("db tags query passed", zap.Any("parentPath", parentPath), zap.Int("resultCount", len(tags)))
	return tags, nil
}

func (db *DigikamSQLDatabase) Ratings() []float64 {
	return []float64{0, 1, 2, 3, 4, 5}
}

func (db *DigikamSQLDatabase) Close() error {
	zap.L().Debug("db close")
	return db.db.Close()
}

// tagIDSubquery accepts a tag and produces a query and the parameters
// associated with that query in order to get the ID of the specified tag
func tagIDSubquery(t types.Tag) (string, []any, error) {
	if len(t.Path) == 0 {
		return "", nil, fmt.Errorf("can't produce a tag subquery for a tag with an empty path")
	}

	parameters := make([]any, 0, len(t.Path))

	// We start with the root tags which have a parent of 0, Then we keep
	// nesting sub-queries till we have built of a query that will find the ID
	// of the specified tag based it's path
	q := "0"
	for _, name := range t.Path {
		q = "(SELECT id FROM Tags WHERE name=? AND pid=" + q + ")"
		parameters = append([]any{name}, parameters...)
	}
	return q, parameters, nil
}

// addCountToQuery modifies the query so as to also return the number of results
// in the query so we are able to create slices with the correct capacity so we
// don't run into inefficient resizes as we add every new result to the slice.
func addCountToQuery(q string) string {
	return "WITH results_before_count AS ( \n\n" + q + "\n\n) SELECT (SELECT COUNT() from results_before_count) as count, * FROM results_before_count"
}
