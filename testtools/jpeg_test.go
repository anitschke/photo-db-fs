package testtools

import (
	"io/fs"
	"testing"

	"github.com/anitschke/photo-db-fs/testtools/mocks"
	"github.com/stretchr/testify/assert"
)

func TestIsJpeg(t *testing.T) {
	assert := assert.New(t)

	assert.True(IsJpeg("./test_photos/valid.jpeg"))
	assert.True(IsJpeg("./test_photos/linkToValid.jpeg"))

	assert.False(IsJpeg("./test_photos/emptyFile.jpeg"))
	assert.False(IsJpeg("./test_photos/linkToEmptyFile.jpeg"))

	assert.False(IsJpeg("./jpeg_test.go"))
}

func TestVerifyJpegAreValid_AllValid(t *testing.T) {
	fileInfos := []FileInfo{
		{
			Path: "./test_photos/valid.jpeg",
			Mode: 0,
		},
		{
			Path:       "./test_photos/linkToValid.jpeg",
			Mode:       fs.ModeSymlink,
			LinkTarget: "valid.jpeg",
		},
		{
			Path: "./jpeg_test.go",
			Mode: 0,
		},
	}

	VerifyJpegAreValid(t, fileInfos)
}

func TestVerifyJpegAreValid_SomeInvalid(t *testing.T) {
	fileInfos := []FileInfo{
		{
			Path: "./test_photos/valid.jpeg",
			Mode: 0,
		},
		{
			Path:       "./test_photos/linkToValid.jpeg",
			Mode:       fs.ModeSymlink,
			LinkTarget: "valid.jpeg",
		},
		{
			Path: "./test_photos/emptyFile.jpeg",
			Mode: 0,
		},
		{
			Path:       "./test_photos/linkToEmptyFile.jpeg",
			Mode:       fs.ModeSymlink,
			LinkTarget: "emptyFile.jpeg",
		},
		{
			Path: "./jpeg_test.go",
			Mode: 0,
		},
	}

	mockT := mocks.NewTestingT(t)
	diag := "%q is not a valid JPEG file"
	mockT.On("Errorf", diag, "./test_photos/emptyFile.jpeg").Once()
	mockT.On("Errorf", diag, "./test_photos/linkToEmptyFile.jpeg").Once()

	VerifyJpegAreValid(mockT, fileInfos)
}
