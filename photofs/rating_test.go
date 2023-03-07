package photofs

// cSpell:words Aand

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/anitschke/photo-db-fs/db"
	"github.com/anitschke/photo-db-fs/db/mocks"
	"github.com/anitschke/photo-db-fs/testtools"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func rootRatingInode(ctx context.Context, db db.DB) (fs.InodeEmbedder, error) {
	n := ratingsParentNode{db: db}
	return n.INode(ctx)
}

func nestedRatingInode(ctx context.Context, db db.DB, baseSelector types.Selector) (fs.InodeEmbedder, error) {
	n := ratingsParentNode{db: db, baseSelector: baseSelector}
	return n.INode(ctx)
}

// WARNING when there are bugs in these tests they tend to deadlock
// even with a timeout specified in the test runner. This seems to happen most
// if the tagFS needs to make a call out to the mock database and that call
// hasn't been defined yet, so it feels like it is some issue where the mock is
// erroring in a way that the tagFS isn't handling correctly and as a result we
// error out during a syscall or something and deadlock.

func TestRatingFS(t *testing.T) {
	assert := assert.New(t)

	mockDB := mocks.NewDB(t)
	mockDB.On("Ratings").Return(
		[]float64{1, 2, 3},
		nil,
	).Once()

	ctx := context.Background()
	ratingNode, err := rootRatingInode(ctx, mockDB)
	assert.NotNil(ratingNode)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, ratingNode)
	assert.Nil(err)
	serverDoneWG := sync.WaitGroup{}
	serverDoneWG.Add(1)
	go func() {
		server.Wait()
		serverDoneWG.Done()
	}()

	defer func() {
		err := server.Unmount()
		assert.Nil(err)
		serverDoneWG.Wait()

		// If we don't walk the DB then we should never need to query about photos
		mockDB.AssertNotCalled(t, "Photos", mock.Anything, mock.Anything)
	}()
}

func TestRatingFS_WalkRatings(t *testing.T) {
	assert := assert.New(t)

	mockDB := mocks.NewDB(t)

	// To make this test simpler we will only deal with walking ratings in this
	// test and not photos. We can walk photos in a different test.
	mockDB.On("Photos", mock.Anything, mock.Anything).Return(
		[]types.Photo{},
		nil,
	)

	mockDB.On("Ratings").Return(
		[]float64{1, 2, 3},
		nil,
	).Once()

	ctx := context.Background()
	ratingNode, err := rootRatingInode(ctx, mockDB)
	assert.NotNil(ratingNode)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, ratingNode)
	assert.Nil(err)
	serverDoneWG := sync.WaitGroup{}
	serverDoneWG.Add(1)
	go func() {
		server.Wait()
		serverDoneWG.Done()
	}()

	defer func() {
		err := server.Unmount()
		assert.Nil(err)
		serverDoneWG.Wait()
	}()

	actTreeInfo, err := testtools.Walk(mountPoint)
	assert.Nil(err)
	testtools.ToGoldFileFormat(actTreeInfo, mountPoint, "NO_LIBRARY_NEEDED_SINCE_NO_PHOTOS")
	updateGold := false
	expTreeInfo := testtools.GetOrUpdateGoldFile("./"+t.Name()+"_GoldTree.json", actTreeInfo, updateGold)
	assert.ElementsMatch(actTreeInfo, expTreeInfo)
}

func TestRatingFS_WalkRatings_DBHasNoRatings(t *testing.T) {
	assert := assert.New(t)

	mockDB := mocks.NewDB(t)

	mockDB.On("Ratings").Return(
		[]float64{},
		nil,
	).Once()

	ctx := context.Background()
	ratingNode, err := rootRatingInode(ctx, mockDB)
	assert.NotNil(ratingNode)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, ratingNode)
	assert.Nil(err)
	serverDoneWG := sync.WaitGroup{}
	serverDoneWG.Add(1)
	go func() {
		server.Wait()
		serverDoneWG.Done()
	}()

	defer func() {
		err := server.Unmount()
		assert.Nil(err)
		serverDoneWG.Wait()
	}()

	actTreeInfo, err := testtools.Walk(mountPoint)
	assert.Nil(err)
	testtools.ToGoldFileFormat(actTreeInfo, mountPoint, "NO_LIBRARY_NEEDED_SINCE_NO_PHOTOS")
	updateGold := false
	expTreeInfo := testtools.GetOrUpdateGoldFile("./"+t.Name()+"_GoldTree.json", actTreeInfo, updateGold)
	assert.ElementsMatch(actTreeInfo, expTreeInfo)
}

func TestRatingFS_WalkPhotos(t *testing.T) {
	assert := assert.New(t)

	mockDB := mocks.NewDB(t)

	mockDB.On("Ratings").Return(
		[]float64{1, 2, 3},
		nil,
	).Once()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	libraryRoot := filepath.Join(wd, "..", "test-resources", "photos", "basic")

	rated1_1 := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_00626.jpg"),
		ID:   "rated1_1",
	}
	rated1_2 := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_00896.jpg"),
		ID:   "rated1_2",
	}

	rated2_1 := types.Photo{
		Path: filepath.Join(libraryRoot, "album2", "DSC_0196.jpg"),
		ID:   "rated2_1",
	}
	rated2_2 := types.Photo{
		Path: filepath.Join(libraryRoot, "album2", "DSC_0340_BW.jpg"),
		ID:   "rated2_2",
	}

	rated3 := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_03476.jpg"),
		ID:   "rated3",
	}

	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.HasRating{Operator: types.Equal, Rating: 1}}).Return(
		[]types.Photo{
			rated1_1,
			rated1_2,
		},
		nil,
	).Once()
	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.HasRating{Operator: types.Equal, Rating: 2}}).Return(
		[]types.Photo{
			rated2_1,
			rated2_2,
		},
		nil,
	).Once()
	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.HasRating{Operator: types.Equal, Rating: 3}}).Return(
		[]types.Photo{
			rated3,
		},
		nil,
	).Once()

	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.HasRating{Operator: types.GreaterThanOrEqual, Rating: 1}}).Return(
		[]types.Photo{
			rated1_1,
			rated1_2,
			rated2_1,
			rated2_2,
			rated3,
		},
		nil,
	).Once()
	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.HasRating{Operator: types.GreaterThanOrEqual, Rating: 2}}).Return(
		[]types.Photo{
			rated2_1,
			rated2_2,
			rated3,
		},
		nil,
	).Once()

	ctx := context.Background()
	ratingNode, err := rootRatingInode(ctx, mockDB)
	assert.NotNil(ratingNode)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, ratingNode)
	assert.Nil(err)
	serverDoneWG := sync.WaitGroup{}
	serverDoneWG.Add(1)
	go func() {
		server.Wait()
		serverDoneWG.Done()
	}()

	defer func() {
		err := server.Unmount()
		assert.Nil(err)
		serverDoneWG.Wait()
	}()

	actTreeInfo, err := testtools.Walk(mountPoint)
	assert.Nil(err)

	testtools.VerifyJpegAreValid(t, actTreeInfo)

	testtools.ToGoldFileFormat(actTreeInfo, mountPoint, libraryRoot)
	updateGold := false
	expTreeInfo := testtools.GetOrUpdateGoldFile("./"+t.Name()+"_GoldTree.json", actTreeInfo, updateGold)
	assert.ElementsMatch(actTreeInfo, expTreeInfo)
}

func TestRatingFS_WalkPhotos_NestedUnderAnotherFS(t *testing.T) {
	assert := assert.New(t)

	mockDB := mocks.NewDB(t)

	mockDB.On("Ratings").Return(
		[]float64{1, 2, 3},
		nil,
	).Once()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	libraryRoot := filepath.Join(wd, "..", "test-resources", "photos", "basic")

	rated1_1 := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_00626.jpg"),
		ID:   "rated1_1",
	}
	rated1_2 := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_00896.jpg"),
		ID:   "rated1_2",
	}

	rated2_1 := types.Photo{
		Path: filepath.Join(libraryRoot, "album2", "DSC_0196.jpg"),
		ID:   "rated2_1",
	}
	rated2_2 := types.Photo{
		Path: filepath.Join(libraryRoot, "album2", "DSC_0340_BW.jpg"),
		ID:   "rated2_2",
	}

	rated3 := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_03476.jpg"),
		ID:   "rated3",
	}

	baseSelector := types.HasTag{Tag: makeTag("myTag")}

	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.And{Operands: []types.Selector{baseSelector, types.HasRating{Operator: types.Equal, Rating: 1}}}}).Return(
		[]types.Photo{
			rated1_1,
			rated1_2,
		},
		nil,
	).Once()
	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.And{Operands: []types.Selector{baseSelector, types.HasRating{Operator: types.Equal, Rating: 2}}}}).Return(
		[]types.Photo{
			rated2_1,
			rated2_2,
		},
		nil,
	).Once()
	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.And{Operands: []types.Selector{baseSelector, types.HasRating{Operator: types.Equal, Rating: 3}}}}).Return(
		[]types.Photo{
			rated3,
		},
		nil,
	).Once()

	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.And{Operands: []types.Selector{baseSelector, types.HasRating{Operator: types.GreaterThanOrEqual, Rating: 1}}}}).Return(
		[]types.Photo{
			rated1_1,
			rated1_2,
			rated2_1,
			rated2_2,
			rated3,
		},
		nil,
	).Once()
	mockDB.On("Photos", mock.Anything, types.Query{Selector: types.And{Operands: []types.Selector{baseSelector, types.HasRating{Operator: types.GreaterThanOrEqual, Rating: 2}}}}).Return(
		[]types.Photo{
			rated2_1,
			rated2_2,
			rated3,
		},
		nil,
	).Once()

	ctx := context.Background()
	ratingNode, err := nestedRatingInode(ctx, mockDB, baseSelector)
	assert.NotNil(ratingNode)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, ratingNode)
	assert.Nil(err)
	serverDoneWG := sync.WaitGroup{}
	serverDoneWG.Add(1)
	go func() {
		server.Wait()
		serverDoneWG.Done()
	}()

	defer func() {
		err := server.Unmount()
		assert.Nil(err)
		serverDoneWG.Wait()
	}()

	actTreeInfo, err := testtools.Walk(mountPoint)
	assert.Nil(err)

	testtools.VerifyJpegAreValid(t, actTreeInfo)

	testtools.ToGoldFileFormat(actTreeInfo, mountPoint, libraryRoot)
	updateGold := false
	expTreeInfo := testtools.GetOrUpdateGoldFile("./"+t.Name()+"_GoldTree.json", actTreeInfo, updateGold)
	assert.ElementsMatch(actTreeInfo, expTreeInfo)
}
