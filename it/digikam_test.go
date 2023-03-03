package integrationtests

import (
	"context"
	"encoding/json"
	"os"
	"strings"
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

	// Loop through all of the photos and make sure we can read them as valid jpeg
	actJPEGFileCount := 0
	for _, f := range actTreeInfo {
		if f.Mode != os.ModeDir {
			assert.True(testtools.IsJpeg(f.Path))
			actJPEGFileCount++
		}
	}
	expJPEGFileCount := 46
	assert.Equal(actJPEGFileCount, expJPEGFileCount)

	// Before we compare against or save the gold file we need to rip the mount
	// point out of the path since that changes every time we run the test.
	for i := 0; i < len(actTreeInfo); i++ {
		actTreeInfo[i].Path = strings.Replace(actTreeInfo[i].Path, mountPoint, "$MOUNT_POINT", 1)
		actTreeInfo[i].LinkTarget = strings.Replace(actTreeInfo[i].LinkTarget, libraryRoot, "$LIBRARY_ROOT", 1)
	}

	updateGold := false
	expTreeInfo := getOrUpdateGoldFile("./digikamGoldTree.json", actTreeInfo, updateGold)

	assert.ElementsMatch(actTreeInfo, expTreeInfo)
}

func getOrUpdateGoldFile(path string, act []testtools.FileInfo, update bool) []testtools.FileInfo {

	if update {
		b, err := json.MarshalIndent(act, "", "    ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(path, b, os.ModePerm); err != nil {
			panic(err)
		}

		panic("update flag should turned off when test is submitted")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	exp := []testtools.FileInfo{}
	if err := json.Unmarshal(b, &exp); err != nil {
		panic(err)
	}
	return exp
}
