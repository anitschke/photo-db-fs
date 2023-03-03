package testtools

import (
	"os"
	"path/filepath"
)

func MakeDirInfo(parts ...string) FileInfo {
	return FileInfo{
		Path: filepath.Join(parts...),
		Mode: os.ModeDir,
	}
}

func MakeSymlinkInfo(path []string, target string) FileInfo {
	return FileInfo{
		Path:       filepath.Join(path...),
		Mode:       os.ModeSymlink,
		LinkTarget: target,
	}
}
