package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueStableName(t *testing.T) {
	{
		p := Photo{
			Path: "path/to/my/photo.jpg",
			ID:   "1234",
		}
		assert.Equal(t, p.UniqueStableName(), "1234.jpg")
	}

	{
		p := Photo{
			Path: "path/to/my/other_photo.png",
			ID:   "789",
		}
		assert.Equal(t, p.UniqueStableName(), "789.png")
	}

	{
		p := Photo{
			Path: "path/to/my/photoWithNoExt",
			ID:   "8765",
		}
		assert.Equal(t, p.UniqueStableName(), "8765")
	}
}

func TestTagName(t *testing.T) {
	{
		tag := Tag{Path: []string{}}
		assert.Equal(t, tag.Name(), "")
	}

	{
		tag := Tag{Path: []string{"MyTag"}}
		assert.Equal(t, tag.Name(), "MyTag")
	}

	{
		tag := Tag{Path: []string{"path", "to", "a", "tag"}}
		assert.Equal(t, tag.Name(), "tag")
	}
}
