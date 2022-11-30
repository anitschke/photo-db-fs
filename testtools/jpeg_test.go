package testtools

import (
	"testing"

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
