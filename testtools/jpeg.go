package testtools

import (
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
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

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
}

// VerifyJpegAreValid verifies that all of the JPEG files in the provided file
// infos are valid JPEG files.
func VerifyJpegAreValid(t TestingT, fileInfos []FileInfo) {
	for _, f := range fileInfos {
		ext := strings.ToLower(filepath.Ext(f.Path))
		if f.Mode != os.ModeDir && (ext == ".jpg" || ext == ".jpeg") {
			valid := IsJpeg(f.Path)
			if !valid {
				t.Errorf("%q is not a valid JPEG file", f.Path)
			}
		}
	}
}
