package integrationtests

import (
	"context"
	"sync"
	"testing"

	"github.com/anitschke/photo-db-fs/db"
	_ "github.com/anitschke/photo-db-fs/db/digikam"
	"github.com/anitschke/photo-db-fs/photofs"
	digikamtestresources "github.com/anitschke/photo-db-fs/test-resources/digikam"
	"github.com/anitschke/photo-db-fs/testtools"
	"github.com/anitschke/photo-db-fs/types"
	"github.com/stretchr/testify/assert"
)

func TestDigikamIntegration(t *testing.T) {
	assert := assert.New(t)

	testDB, libraryRoot, cleanup, err := digikamtestresources.PrepareBasicDB()
	assert.Nil(err)
	defer cleanup()

	mountPoint, cleanup, err := testtools.MountPoint()
	assert.Nil(err)
	defer cleanup()
	ctx := context.Background()

	db, err := db.New("digikam-sqlite", testDB)
	assert.Nil(err)

	queries := []types.NamedQuery{
		{
			Name: "Rafter1OrKayaker",
			Query: types.Query{
				Selector: types.Or{
					Operands: []types.Selector{
						types.HasTag{Tag: types.Tag{Path: []string{"People", "rafter1"}}},
						types.HasTag{Tag: types.Tag{Path: []string{"People", "kayaker"}}},
					},
				},
			},
		},
		{
			Name: "KayakingOrSkiing",
			Query: types.Query{
				Selector: types.Or{
					Operands: []types.Selector{
						types.HasTag{Tag: types.Tag{Path: []string{"activity", "watersports", "kayaking"}}},
						types.HasTag{Tag: types.Tag{Path: []string{"activity", "skiing"}}},
					},
				},
			},
		},
	}

	server, err := photofs.Mount(ctx, mountPoint, db, queries)
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
	expTreeInfo := testtools.GetOrUpdateGoldFile("./digikamGoldTree.json", actTreeInfo, updateGold)
	assert.ElementsMatch(actTreeInfo, expTreeInfo)
}
