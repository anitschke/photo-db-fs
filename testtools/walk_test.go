package testtools

import (
	"io/fs"
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
