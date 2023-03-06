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

func rootQueriesInode(ctx context.Context, db db.DB, queries []types.NamedQuery) (fs.InodeEmbedder, error) {
	n := rootQueriesNode{db: db, queries: queries}
	return n.INode(ctx)
}

// WARNING when there are bugs in these tests they tend to deadlock
// even with a timeout specified in the test runner. This seems to happen most
// if the tagFS needs to make a call out to the mock database and that call
// hasn't been defined yet, so it feels like it is some issue where the mock is
// erroring in a way that the tagFS isn't handling correctly and as a result we
// error out during a syscall or something and deadlock.

func TestQueriesFS(t *testing.T) {
	assert := assert.New(t)

	mockDB := mocks.NewDB(t)
	q := []types.NamedQuery{
		{
			Name: "query1",
		},
		{
			Name: "query2",
		},
		{
			Name: "query3",
		},
	}

	ctx := context.Background()
	queriesRoot, err := rootQueriesInode(ctx, mockDB, q)
	assert.NotNil(queriesRoot)
	assert.Nil(err)

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()

	server, err := testtools.MountTestFs(mountPoint, queriesRoot)
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

		// We shouldn't need to call the db for anything since all we should
		// need when mounting the root is the names of the queries, we shouldn't
		// need to perform them until we want to look inside one of those
		// folders.
		mockDB.AssertNotCalled(t, "Photos", mock.Anything, mock.Anything)
		mockDB.AssertNotCalled(t, "RootTags", mock.Anything)
		mockDB.AssertNotCalled(t, "ChildrenTags", mock.Anything, mock.Anything)
	}()
}

func TestQueriesFS_WalkPhotos(t *testing.T) {
	assert := assert.New(t)

	mockDB := mocks.NewDB(t)

	// Even though we don't really care what the selectors are we need to give
	// each query a real different selector. This is because when we pass the
	// query to the DB we pass the Query not the NamedQuery. So if we don't give
	// a different selector to each query the Query portion of the NamedQuery
	// object is an identical zero Query and our mocking infrastructure can't
	// tell them apart so doesn't know which of the results to return.
	q := []types.NamedQuery{
		{
			Name:  "query1",
			Query: types.Query{Selector: types.HasTag{Tag: makeTag("query1")}},
		},
		{
			Name:  "query2",
			Query: types.Query{Selector: types.HasTag{Tag: makeTag("query2")}},
		},
		{
			Name:  "query3",
			Query: types.Query{Selector: types.HasTag{Tag: makeTag("query3")}},
		},
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	libraryRoot := "../test-resources/photos/basic"

	photo_q1 := types.Photo{
		Path: filepath.Join(wd, libraryRoot+"/album1/GRAND_00626.jpg"),
		ID:   "photo_q1",
	}
	photo_q2 := types.Photo{
		Path: filepath.Join(wd, libraryRoot+"/album1/GRAND_00896.jpg"),
		ID:   "photo_q2",
	}

	photo_q3 := types.Photo{
		Path: filepath.Join(wd, libraryRoot+"/album2/DSC_0196.jpg"),
		ID:   "photo_q3",
	}
	photo_q1_q2 := types.Photo{
		Path: filepath.Join(wd, libraryRoot+"/album2/DSC_0340_BW.jpg"),
		ID:   "photo_q1_q2",
	}

	photo_q1_q3 := types.Photo{
		Path: filepath.Join(wd, libraryRoot+"/album1/GRAND_03476.jpg"),
		ID:   "photo_q1_q3",
	}

	mockDB.On("Photos", mock.Anything, q[0].Query).Return(
		[]types.Photo{
			photo_q1,
			photo_q1_q2,
			photo_q1_q3,
		},
		nil,
	).Once()
	mockDB.On("Photos", mock.Anything, q[1].Query).Return(
		[]types.Photo{
			photo_q2,
			photo_q1_q2,
		},
		nil,
	).Once()
	mockDB.On("Photos", mock.Anything, q[2].Query).Return(
		[]types.Photo{
			photo_q3,
			photo_q1_q3,
		},
		nil,
	).Once()

	ctx := context.Background()
	tagRoot, err := rootQueriesInode(ctx, mockDB, q)
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
