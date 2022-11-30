package testtools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMountPoint(t *testing.T) {
	assert := assert.New(t)

	path, cleanup, err := MountPoint()
	assert.NotEqual(path, "")
	assert.NotNil(cleanup)
	assert.Nil(err)

	assert.DirExists(path)

	cleanup()

	assert.NoDirExists(path)
}
