package config

import (
	"flag"

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
		//xxx parse config file
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
