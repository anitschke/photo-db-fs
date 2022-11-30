package testtools

import (
	"io/fs"
	"os"
	"path/filepath"
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
