package digikamtestresources

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/anitschke/photo-db-fs/utils"
	_ "github.com/mattn/go-sqlite3"
)

// PrepareDB takes the test db that is checked into git prepares it for testing
//
// The test db that is checked into git contains an album root that is an
// absolute path to where it is on my machine. I did some looking around I and
// don't see a way digikam can be setup to use something like an environment
// variable for that path (and even if we could then the symlinking that our
// FUSE file system wouldn't work since that can't handle environment variable).
// So what we are doing here is making a temp copy of the db and then digging
// into the db and modifying the album root so it points to the absolute path of
// where this repository is actually located.
func PrepareDB(srcDBFile string, libraryRoot string) (string, string, func(), error) {
	source, err := os.Open(srcDBFile)
	if err != nil {
		return "", "", nil, err
	}
	defer source.Close()

	dbFile, err := os.CreateTemp("", "photo-db-fs_TEMP_DIGIKAM_DB")
	if err != nil {
		return "", "", nil, err
	}
	defer utils.CloseAndPanicOnError(dbFile)
	cleanup := func() {
		err := os.Remove(dbFile.Name())
		if err != nil {
			panic(fmt.Errorf("failed to cleanup %q: %w", dbFile.Name(), err))
		}
	}

	_, err = io.Copy(dbFile, source)
	if err != nil {
		cleanup()
		return "", "", nil, err
	}

	db, err := sql.Open("sqlite3", "file:"+dbFile.Name())
	if err != nil {
		cleanup()
		return "", "", nil, err
	}

	defer utils.CloseAndPanicOnError(db)
	if _, err := db.Exec(`UPDATE AlbumRoots SET specificPath = "` + libraryRoot + `" WHERE id = 1`); err != nil {
		cleanup()
		return "", "", nil, err
	}

	return dbFile.Name(), libraryRoot, cleanup, nil
}

func PrepareBasicDB() (string, string, func(), error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", "", nil, fmt.Errorf("failed to get path to db")
	}
	currentDir := filepath.Dir(currentFile)

	dbFile := filepath.Join(currentDir, "basic", "db", "digikam4.db")
	libraryRoot := filepath.Join(currentDir, "basic", "photos")
	return PrepareDB(dbFile, libraryRoot)
}

func PrepareSQLInjectionDB() (string, string, func(), error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", "", nil, fmt.Errorf("failed to get path to db")
	}
	currentDir := filepath.Dir(currentFile)

	dbFile := filepath.Join(currentDir, "sql_injection", "db", "digikam4.db")
	libraryRoot := filepath.Join(currentDir, "sql_injection", "photos")
	return PrepareDB(dbFile, libraryRoot)
}
