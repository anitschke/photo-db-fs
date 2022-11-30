package types

import "path/filepath"

type Photo struct {

	// Absolute path to the photo
	Path string

	// ID should be a unique ID to all photos in that may be returned by a photo
	// database.
	//
	// The purpose of this ID is that frequently in photo libraries there can be
	// multiple photos with the same names, but since we are flattening multiple
	// directory hierarchies into a single folder it leaves the real possibility
	// of name collisions. This means we need a unique identifier to be used for
	// photos so that even if photos are named the same thing in the folder they
	// exist in we can give them a unique name in the tag folder we are
	// flattening the photos into.
	//
	// As such this ID will be used for the name of the FUSE file system.
	//
	// Note that care must be given for this ID to actually be unique. In my
	// initial implementation I was using the Digikam uniqueHash, which is a md5
	// hash of the first and last 100 kb of the photo (
	// https://github.com/KDE/digikam/blob/33d0457e20adda97c003f3dee652a1749406ff9f/core/libs/dimg/loaders/dimgloader.cpp#L333
	// ). This means that if a photo is copied somewhere we can have two files
	// with the same hash. In this case we will just ignore once of the results,
	// since both photos are the same it is ok to drop one of them from the
	// results, so we sort of have a nice built in duplicate detection and
	// rejection. We could instead use the photo ID, but for some reason I
	// prefer using the md5 hash based naming since it is a nice stable way of
	// identifying an photo even if it moves around in a way that can't be
	// detected by the digikam DB.
	ID string
}

// UniqueStableName generates a unique and stable name for a photo using the
// unique ID provided by the DB.
func (p Photo) UniqueStableName() string {
	return p.ID + filepath.Ext(p.Path)
}

type Tag struct {
	Path []string
}

func (t Tag) Name() string {
	if t.Path == nil || len(t.Path) == 0 {
		return ""
	}
	return t.Path[len(t.Path)-1]
}
