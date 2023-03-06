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

func rootTagInode(ctx context.Context, db db.DB) (fs.InodeEmbedder, error) {
	n := rootTagsNode{db: db}
	return n.INode(ctx)
}

func makeTag(path ...string) types.Tag {
	return types.Tag{
		Path: path,
	}
}

// WARNING when there are bugs in these tests they tend to deadlock
// even with a timeout specified in the test runner. This seems to happen most
// if the tagFS needs to make a call out to the mock database and that call
// hasn't been defined yet, so it feels like it is some issue where the mock is
// erroring in a way that the tagFS isn't handling correctly and as a result we
// error out during a syscall or something and deadlock.

func TestTagFS(t *testing.T) {
	assert := assert.New(t)

	tagDB := mocks.NewDB(t)
	tagDB.On("RootTags", mock.Anything).Return(
		[]types.Tag{
			makeTag("a"),
			makeTag("b"),
			makeTag("c"),
		},
		nil,
	).Once()

	ctx := context.Background()
	tagRoot, err := rootTagInode(ctx, tagDB)
	assert.NotNil(tagRoot)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, tagRoot)
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

		// If we don't walk the DB then we should never need to ask for children
		// tags of these root tags
		tagDB.AssertNotCalled(t, "ChildrenTags", mock.Anything, mock.Anything)
	}()
}

func TestTagFS_WalkTags(t *testing.T) {
	assert := assert.New(t)

	tagDB := mocks.NewDB(t)

	// To make this test simpler we will only deal with walking tags in this
	// test and not photos. We can walk photos in a different test.
	tagDB.On("Photos", mock.Anything, mock.Anything).Return(
		[]types.Photo{},
		nil,
	)

	// Make the test simpler by making the DB say that it doesn't have any ratings
	tagDB.On("Ratings").Return(
		[]float64{},
		nil,
	)

	tagDB.On("RootTags", mock.Anything).Return(
		[]types.Tag{
			makeTag("a"),
			makeTag("b"),
			makeTag("c"),
		},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("a")).Return(
		[]types.Tag{
			makeTag("a", "a"),
			makeTag("a", "b"),
			makeTag("a", "c"),
		},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("a", "a")).Return(
		[]types.Tag{},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("a", "b")).Return(
		[]types.Tag{},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("a", "c")).Return(
		[]types.Tag{},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("b")).Return(
		[]types.Tag{
			makeTag("b", "a"),
			makeTag("b", "b"),
			makeTag("b", "c"),
		},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("b", "a")).Return(
		[]types.Tag{},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("b", "b")).Return(
		[]types.Tag{},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("b", "c")).Return(
		[]types.Tag{},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("c")).Return(
		[]types.Tag{},
		nil,
	).Once()

	ctx := context.Background()
	tagRoot, err := rootTagInode(ctx, tagDB)
	assert.NotNil(tagRoot)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, tagRoot)
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

func TestTagFS_WalkPhotos(t *testing.T) {
	assert := assert.New(t)

	tagDB := mocks.NewDB(t)

	tagDB.On("RootTags", mock.Anything).Return(
		[]types.Tag{
			makeTag("a"),
		},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("a")).Return(
		[]types.Tag{
			makeTag("a", "a"),
		},
		nil,
	).Once()
	tagDB.On("ChildrenTags", mock.Anything, makeTag("a", "a")).Return(
		[]types.Tag{},
		nil,
	).Once()

	// Make the test simpler by making the DB say that it doesn't have any ratings
	tagDB.On("Ratings").Return(
		[]float64{},
		nil,
	)

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	libraryRoot := filepath.Join(wd, "..", "test-resources", "photos", "basic")

	taggedByA1 := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_00626.jpg"),
		ID:   "taggedByA1",
	}
	taggedByA2 := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_00896.jpg"),
		ID:   "taggedByA2",
	}

	taggedByAA1 := types.Photo{
		Path: filepath.Join(libraryRoot, "album2", "DSC_0196.jpg"),
		ID:   "taggedByAA1",
	}
	taggedByAA2 := types.Photo{
		Path: filepath.Join(libraryRoot, "album2", "DSC_0340_BW.jpg"),
		ID:   "taggedByAA2",
	}

	taggedByAandAA := types.Photo{
		Path: filepath.Join(libraryRoot, "album1", "GRAND_03476.jpg"),
		ID:   "taggedByAandAA",
	}

	tagDB.On("Photos", mock.Anything, types.Query{Selector: types.HasTag{Tag: makeTag("a")}}).Return(
		[]types.Photo{
			taggedByA1,
			taggedByA2,
			taggedByAandAA,
		},
		nil,
	).Once()
	tagDB.On("Photos", mock.Anything, types.Query{Selector: types.HasTag{Tag: makeTag("a", "a")}}).Return(
		[]types.Photo{
			taggedByAA1,
			taggedByAA2,
			taggedByAandAA,
		},
		nil,
	).Once()

	ctx := context.Background()
	tagRoot, err := rootTagInode(ctx, tagDB)
	assert.NotNil(tagRoot)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, tagRoot)
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
