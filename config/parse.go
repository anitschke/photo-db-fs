package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/anitschke/photo-db-fs/types"
)

var configFileFlag = flag.String("config-file", "", "photo-db-fs config file")

var mountPointFlag = flag.String("mount-point", "", "location where photo-db-fs file system will be mounted")
var dbTypeFlag = flag.String("db-type", "", "type of photo database photo-db-fs will use for querying")
var dbSource = flag.String("db-source", "", "source of the database photo-db-fs will use for querying, ie for local databases this is the path to the database")

var logLevelFlag = flag.String("log-level", "", "debugging logging level")

func Parse() (types.Config, error) {
	flag.Parse()

	var config types.Config

	if *configFileFlag != "" {
		fileBytes, err := os.ReadFile(*configFileFlag)
		if err != nil {
			return types.Config{}, fmt.Errorf("failed to read config file %q: %w", *configFileFlag, err)
		}

		if err := json.Unmarshal(fileBytes, &config); err != nil {
			return types.Config{}, fmt.Errorf("failed to parse config file %q as JSON: %w", *configFileFlag, err)
		}
	}

	if *mountPointFlag != "" {
		config.MountPoint = *mountPointFlag
	}
	if *dbTypeFlag != "" {
		config.DB.Type = *dbTypeFlag
	}
	if *dbSource != "" {
		config.DB.Source = *dbSource
	}
	if *logLevelFlag != "" {
		config.LogLevel = *logLevelFlag
	}

	return config, nil
}
