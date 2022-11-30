package digikam

import (
	"context"
	"testing"

	digikamtestresources "github.com/anitschke/photo-db-fs/test-resources/digikam"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/stretchr/testify/assert"
)

// When I initially wrote the driver for getting data from digiKam db I was lazy
// and just concatenated data like tag names into the query. This opens us up to
// sql injection style bugs if the tag names happen to have a single or double
// quote in them.
//
// To fix this any sort of string that comes from the user needs to be passed
// over to sqlite query via a variable/parameter instead of concatenated into
// the query string.
//
// This is a test to lock down that we can correctly process data from the user
// that might cause one of these sql injection style bugs.

func TestDigikamSqlInjectionSqliteDatabase(t *testing.T) {
	assert := assert.New(t)

	testDB, _, cleanup, err := digikamtestresources.PrepareSQLInjectionDB()
	assert.Nil(err)
	defer cleanup()

	db, err := NewDigikamSqliteDatabase(testDB)
	assert.NotNil(db)
	assert.Nil(err)

	err = db.Close()
	assert.Nil(err)
}

func TestDigikamSqlInjectionSqliteDatabase_RootTags(t *testing.T) {
	assert := assert.New(t)

	testDB, _, cleanup, err := digikamtestresources.PrepareSQLInjectionDB()
	assert.Nil(err)
	defer cleanup()

	db, err := NewDigikamSqliteDatabase(testDB)
	assert.Nil(err)
	defer func() {
		err = db.Close()
		assert.Nil(err)
	}()

	ctx := context.Background()
	actTags, err := db.RootTags(ctx)

	expTags := []types.Tag{
		{
			Path: []string{"_Digikam_Internal_Tags_"},
		},
		{
			Path: []string{`"'; DROP TABLE  Images; `},
		},
	}

	assert.ElementsMatch(actTags, expTags)
}

func TestDigikamSqlInjectionSqliteDatabase_ChildrenTags(t *testing.T) {
	assert := assert.New(t)

	testDB, _, cleanup, err := digikamtestresources.PrepareSQLInjectionDB()
	assert.Nil(err)
	defer cleanup()

	db, err := NewDigikamSqliteDatabase(testDB)
	assert.Nil(err)
	defer func() {
		err = db.Close()
		assert.Nil(err)
	}()

	parentTag := types.Tag{
		Path: []string{`"'; DROP TABLE  Images; `},
	}

	ctx := context.Background()
	actTags, err := db.ChildrenTags(ctx, parentTag)
	assert.Nil(err)

	expTags := []types.Tag{
		{
			Path: []string{`"'; DROP TABLE  Images; `, `"'; DROP TABLE Tags; `},
		},
	}

	assert.ElementsMatch(actTags, expTags)
}

func TestDigikamSqlInjectionSqliteDatabase_Photos(t *testing.T) {
	assert := assert.New(t)

	testDB, libraryRoot, cleanup, err := digikamtestresources.PrepareSQLInjectionDB()
	assert.Nil(err)
	defer cleanup()

	db, err := NewDigikamSqliteDatabase(testDB)
	assert.Nil(err)
	defer func() {
		err = db.Close()
		assert.Nil(err)
	}()

	selector := types.HasTag(
		types.HasTag{
			Tag: types.Tag{
				Path: []string{`"'; DROP TABLE  Images; `},
			},
		},
	)

	q := types.Query{
		Selector: selector,
	}

	ctx := context.Background()
	actPhotos, err := db.Photos(ctx, q)
	assert.Nil(err)

	expPhotos := []types.Photo{
		{
			Path: libraryRoot + "/album1/GRAND_00896.jpg",
			ID:   "de7303f2c490dc1b3fe23b0e17277542",
		},
	}

	assert.ElementsMatch(actPhotos, expPhotos)
}
