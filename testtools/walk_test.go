package testtools

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalk(t *testing.T) {
	actInfo, err := Walk("./walk_test_folder")
	assert.Nil(t, err)

	expInfo := []FileInfo{
		{
			Path: "./walk_test_folder",
			Mode: fs.ModeDir,
		},
		{
			Path: "walk_test_folder/a.txt",
			Mode: 0,
		},
		{
			Path:       "walk_test_folder/aLink.txt",
			Mode:       fs.ModeSymlink,
			LinkTarget: "./a.txt",
		},
		{
			Path: "walk_test_folder/b.txt",
			Mode: 0,
		},
		{
			Path: "walk_test_folder/c.txt",
			Mode: 0,
		},
		{
			Path: "walk_test_folder/child",
			Mode: fs.ModeDir,
		},
		{
			Path: "walk_test_folder/child/aa.txt",
			Mode: 0,
		},
		{
			Path: "walk_test_folder/child/grandchild",
			Mode: fs.ModeDir,
		},
		{
			Path: "walk_test_folder/child/grandchild/aaa.txt",
			Mode: 0,
		},
	}

	assert.ElementsMatch(t, actInfo, expInfo)
}

func TestToGoldFileFormat(t *testing.T) {
	mountPoint := "/path/to/mount/point"
	libraryRoot := "/path/to/library/root"

	actFileInfos := []FileInfo{
		{
			Path: mountPoint + "/one.jpeg",
			Mode: 0,
		},
		{
			Path:       mountPoint + "/two.jpeg",
			Mode:       fs.ModeSymlink,
			LinkTarget: libraryRoot + "/twoInLib.jpeg",
		},
		{
			Path: mountPoint + "/three.jpeg",
			Mode: 0,
		},
		{
			Path:       mountPoint + "/four.jpeg",
			Mode:       fs.ModeSymlink,
			LinkTarget: libraryRoot + "/forInLib.jpeg",
		},
	}

	expFileInfos := []FileInfo{
		{
			Path: "$MOUNT_POINT/one.jpeg",
			Mode: 0,
		},
		{
			Path:       "$MOUNT_POINT/two.jpeg",
			Mode:       fs.ModeSymlink,
			LinkTarget: "$LIBRARY_ROOT/twoInLib.jpeg",
		},
		{
			Path: "$MOUNT_POINT/three.jpeg",
			Mode: 0,
		},
		{
			Path:       "$MOUNT_POINT/four.jpeg",
			Mode:       fs.ModeSymlink,
			LinkTarget: "$LIBRARY_ROOT/forInLib.jpeg",
		},
	}

	ToGoldFileFormat(actFileInfos, mountPoint, libraryRoot)

	assert.Equal(t, actFileInfos, expFileInfos)
}

func TestGetOrUpdateGoldFile_Get(t *testing.T) {

	treeInfoForUpdate := []FileInfo{
		{
			Path: "$MOUNT_POINT/photoForUpdate.jpeg",
			Mode: 0,
		},
	}
	updateGold := false

	treeInfoFromGetOrUpdate := GetOrUpdateGoldFile("./TestGetOrUpdateGoldFile_Get_GoldTree.json", treeInfoForUpdate, updateGold)

	expTreeInfoFromGetOrUpdate := []FileInfo{
		{
			Path: "$MOUNT_POINT/existingPhoto.jpeg",
			Mode: 0,
		},
	}

	assert.Equal(t, treeInfoFromGetOrUpdate, expTreeInfoFromGetOrUpdate)
}

func TestGetOrUpdateGoldFile_Update(t *testing.T) {
	goldFile, err := os.CreateTemp("", "TestGetOrUpdateGoldFile_Update_GoldTree_*.json")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := goldFile.Close(); err != nil {
			panic(err)
		}
	}()

	treeInfoForUpdate := []FileInfo{
		{
			Path: "$MOUNT_POINT/photoForUpdate.jpeg",
			Mode: 0,
		},
	}
	updateGold := true

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code should panic when updating to prevent checking in versions with update set to true")
			}
		}()

		GetOrUpdateGoldFile(goldFile.Name(), treeInfoForUpdate, updateGold)
	}()

	actFileContents, err := io.ReadAll(goldFile)
	assert.NoError(t, err)

	var actTree []FileInfo
	err = json.Unmarshal(actFileContents, &actTree)
	assert.NoError(t, err)

	assert.Equal(t, actTree, treeInfoForUpdate)
}
