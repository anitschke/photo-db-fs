package testtools

import (
	"image/jpeg"
	"os"
)

// IsJpeg is a simple test tool to read in the specified file to see if it is a
// jpeg photo. This is intended to give us good enough confidence that the photo
// at that path had symlinking setup correctly by photofs. As long as we can
// read it as a valid jpeg then the symlinking worked. (we are also verifying
// the link target, this is just one added layer to make sure we can actually
// read the photo too.)
func IsJpeg(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	// If we can read the file without erroring then we know that we have a
	// valid jpeg file
	_, err = jpeg.Decode(f)
	return err == nil
}
