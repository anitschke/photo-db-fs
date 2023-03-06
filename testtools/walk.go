package testtools

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Path       string      `json:"path"`
	Mode       fs.FileMode `json:"mode"`
	LinkTarget string      `json:"linkTarget,omitempty"`
}

func Walk(walkPath string) ([]FileInfo, error) {
	info := make([]FileInfo, 0)
	err := filepath.WalkDir(walkPath, func(path string, d fs.DirEntry, err error) error {

		var linkTarget string
		if d.Type() == fs.ModeSymlink {
			var readErr error
			linkTarget, readErr = os.Readlink(path)
			if readErr != nil {
				return readErr
			}
		}

		info = append(info, FileInfo{
			Path:       path,
			Mode:       d.Type(),
			LinkTarget: linkTarget,
		})
		return nil
	})
	return info, err
}

// ToGoldFileFormat does some processing of the "Walk" results so it may be
// stored as a "gold" file test result file
func ToGoldFileFormat(treeInfo []FileInfo, mountPoint string, libraryRoot string) {
	// Before we compare against or save the gold file we need to rip the mount
	// point out of the path since that changes every time we run the test.
	for i := 0; i < len(treeInfo); i++ {
		treeInfo[i].Path = strings.Replace(treeInfo[i].Path, mountPoint, "$MOUNT_POINT", 1)
		treeInfo[i].LinkTarget = strings.Replace(treeInfo[i].LinkTarget, libraryRoot, "$LIBRARY_ROOT", 1)
	}
}

// GetOrUpdateGoldFile gets an existing gold file or takes the actual file info
// and uses it to overwrite the existing gold file.
func GetOrUpdateGoldFile(path string, act []FileInfo, update bool) []FileInfo {

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

	exp := []FileInfo{}
	if err := json.Unmarshal(b, &exp); err != nil {
		panic(err)
	}
	return exp
}
