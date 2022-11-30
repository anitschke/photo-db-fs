package digikam

import (
	"context"
	"testing"

	"github.com/anitschke/photo-db-fs/db"
	digikamtestresources "github.com/anitschke/photo-db-fs/test-resources/digikam"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/stretchr/testify/assert"
)

func TestDigikamSqliteDatabase(t *testing.T) {
	assert := assert.New(t)

	testDB, _, cleanup, err := digikamtestresources.PrepareBasicDB()
	assert.Nil(err)
	defer cleanup()

	db, err := NewDigikamSqliteDatabase(testDB)
	assert.NotNil(db)
	assert.Nil(err)

	err = db.Close()
	assert.Nil(err)
}

func TestDigikamSqliteDatabase_Registered(t *testing.T) {
	assert := assert.New(t)

	testDB, _, cleanup, err := digikamtestresources.PrepareBasicDB()
	assert.Nil(err)
	defer cleanup()

	db, err := db.New("digikam-sqlite", testDB)
	assert.NotNil(db)
	assert.Nil(err)

	err = db.Close()
	assert.Nil(err)
}

func TestDigikamSqliteDatabase_RootTags(t *testing.T) {
	assert := assert.New(t)

	testDB, _, cleanup, err := digikamtestresources.PrepareBasicDB()
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
			Path: []string{"People"},
		},
		{
			Path: []string{"_Digikam_Internal_Tags_"},
		},
		{
			Path: []string{"activity"},
		},
	}

	assert.ElementsMatch(actTags, expTags)
}

func TestDigikamSqliteDatabase_ChildrenTags(t *testing.T) {
	assert := assert.New(t)

	testDB, _, cleanup, err := digikamtestresources.PrepareBasicDB()
	assert.Nil(err)
	defer cleanup()

	db, err := NewDigikamSqliteDatabase(testDB)
	assert.Nil(err)
	defer func() {
		err = db.Close()
		assert.Nil(err)
	}()

	parentTag := types.Tag{
		Path: []string{"activity"},
	}

	ctx := context.Background()
	actTags, err := db.ChildrenTags(ctx, parentTag)
	assert.Nil(err)

	expTags := []types.Tag{
		{
			Path: []string{"activity", "skiing"},
		},
		{
			Path: []string{"activity", "watersports"},
		},
	}

	assert.ElementsMatch(actTags, expTags)
}

func TestDigikamSqliteDatabase_ChildrenTags_ChildrenOfNonRoot(t *testing.T) {
	assert := assert.New(t)

	testDB, _, cleanup, err := digikamtestresources.PrepareBasicDB()
	assert.Nil(err)
	defer cleanup()

	db, err := NewDigikamSqliteDatabase(testDB)
	assert.Nil(err)
	defer func() {
		err = db.Close()
		assert.Nil(err)
	}()

	parentTag := types.Tag{
		Path: []string{"activity", "watersports"},
	}

	ctx := context.Background()
	actTags, err := db.ChildrenTags(ctx, parentTag)
	assert.Nil(err)

	expTags := []types.Tag{
		{
			Path: []string{"activity", "watersports", "rafting"},
		},
		{
			Path: []string{"activity", "watersports", "kayaking"},
		},
	}

	assert.ElementsMatch(actTags, expTags)
}

func TestDigikamSqliteDatabase_Photos_basic_tag(t *testing.T) {
	assert := assert.New(t)

	testDB, libraryRoot, cleanup, err := digikamtestresources.PrepareBasicDB()
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
				Path: []string{"activity", "skiing"},
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
			Path: libraryRoot + "/album2/DSC_0196.jpg",
			ID:   "d5b701b4043c51007430119971b17ae2",
		},
		{
			Path: libraryRoot + "/album2/DSC_0340_BW.jpg",
			ID:   "17db9d693f682a894fb0ff538dccb972",
		},
		{
			Path: libraryRoot + "/album2/DSC_6603.jpg",
			ID:   "0048360c4b329c9b14925fe2db2a7b34",
		},
	}

	assert.ElementsMatch(actPhotos, expPhotos)
}

func TestDigikamSqliteDatabase_Photos_set_operations(t *testing.T) {
	// Querying for photos allows complex set operations but the logic to build
	// these queries can get a little tricky.

	assert := assert.New(t)

	testDB, libraryRoot, cleanup, err := digikamtestresources.PrepareBasicDB()
	assert.Nil(err)
	defer cleanup()

	db, err := NewDigikamSqliteDatabase(testDB)
	assert.Nil(err)
	defer func() {
		err = db.Close()
		assert.Nil(err)
	}()

	testQuery := func(selector types.Selector, expPhotos []types.Photo) {
		q := types.Query{
			Selector: selector,
		}

		ctx := context.Background()
		actPhotos, err := db.Photos(ctx, q)
		assert.Nil(err)
		assert.ElementsMatch(actPhotos, expPhotos)
	}

	// List of the photos
	watersportsRed := types.Photo{Path: libraryRoot + "/album1/GRAND_00626.jpg", ID: "35f0ac735f2e0f585cac5b918bf98bf3"}
	watersportsGreen := types.Photo{Path: libraryRoot + "/album1/GRAND_00896.jpg", ID: "de7303f2c490dc1b3fe23b0e17277542"}
	watersportsNone := types.Photo{Path: libraryRoot + "/album1/GRAND_01471.jpg", ID: "8c91175a9a7cac20d821835e92091154"}

	skiingRed := types.Photo{Path: libraryRoot + "/album2/DSC_0196.jpg", ID: "d5b701b4043c51007430119971b17ae2"}
	skiingGreen := types.Photo{Path: libraryRoot + "/album2/DSC_6603.jpg", ID: "0048360c4b329c9b14925fe2db2a7b34"}
	skiingNone := types.Photo{Path: libraryRoot + "/album2/DSC_0340_BW.jpg", ID: "17db9d693f682a894fb0ff538dccb972"}

	untaggedRed := types.Photo{Path: libraryRoot + "/album1/GRAND_03331.jpg", ID: "fa1f19e1bc9216e68689acd11044b0ed"}
	untaggedGreen := types.Photo{Path: libraryRoot + "/album1/GRAND_03476.jpg", ID: "f5e76142783d0c7466b4bcc8fcc9afff"}

	// List of basic tag selectors
	hasWatersports := types.HasTag{Tag: types.Tag{Path: []string{"activity", "watersports"}}}
	hasSkiing := types.HasTag{Tag: types.Tag{Path: []string{"activity", "skiing"}}}
	hasRed := types.HasTag{Tag: types.Tag{Path: []string{"_Digikam_Internal_Tags_", "Color Label Red"}}}
	hasGreen := types.HasTag{Tag: types.Tag{Path: []string{"_Digikam_Internal_Tags_", "Color Label Green"}}}

	// First lets just verify that all the normal queries work correctly before
	// we start doing any crazy set operations
	testQuery(hasWatersports, []types.Photo{watersportsRed, watersportsGreen, watersportsNone})
	testQuery(hasSkiing, []types.Photo{skiingRed, skiingGreen, skiingNone})
	testQuery(hasRed, []types.Photo{watersportsRed, skiingRed, untaggedRed})
	testQuery(hasGreen, []types.Photo{watersportsGreen, skiingGreen, untaggedGreen})

	// Now lets do crazy set operations

	// Or
	testQuery(types.Or{Operands: []types.Selector{hasWatersports, hasSkiing}}, []types.Photo{
		watersportsRed, watersportsGreen, watersportsNone,
		skiingRed, skiingGreen, skiingNone})

	testQuery(types.Or{Operands: []types.Selector{hasWatersports, hasSkiing, hasRed, hasGreen}}, []types.Photo{
		watersportsRed, watersportsGreen, watersportsNone,
		skiingRed, skiingGreen, skiingNone,
		untaggedRed,
		untaggedGreen,
	})

	// And
	testQuery(types.And{Operands: []types.Selector{hasWatersports, hasRed}}, []types.Photo{
		watersportsRed})

	testQuery(types.And{Operands: []types.Selector{hasWatersports, hasGreen}}, []types.Photo{
		watersportsGreen})

	// Difference
	testQuery(types.Difference{Starting: hasWatersports, Excluding: hasRed}, []types.Photo{
		watersportsGreen, watersportsNone})

	testQuery(types.Difference{Starting: hasWatersports, Excluding: hasGreen}, []types.Photo{
		watersportsRed, watersportsNone})

	// Many set operations together
	selector := types.Or{Operands: []types.Selector{
		types.And{Operands: []types.Selector{
			hasSkiing,
			hasRed,
		}},
		types.Difference{
			Starting:  hasWatersports,
			Excluding: hasRed,
		},
	}}
	testQuery(selector, []types.Photo{
		skiingRed,
		watersportsGreen,
		watersportsNone,
	})
}
