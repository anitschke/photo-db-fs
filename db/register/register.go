package db

// The expectation is that all DB should register themselves in an init()
// function in their package. Then by including themselves in this file (which
// is included in the main.go) we expose all of the db registrations to the end
// user.

import _ "github.com/anitschke/photo-db-fs/db/digikam"
