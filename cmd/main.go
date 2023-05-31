// Copyright (C) 2023 by Anthony V. Ozdemir
// MIT License (refer to LICENSE)

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	embedding "github.com/anthony-ozdemir/zfse"
	"github.com/anthony-ozdemir/zfse/internal/app"
	"github.com/anthony-ozdemir/zfse/internal/config"
	"github.com/anthony-ozdemir/zfse/internal/helper"
	"github.com/anthony-ozdemir/zfse/internal/path_manager"
)

func main() {
	// Setup Logger
	logger := zap.NewExample()
	zap.ReplaceGlobals(logger)
	defer func() {
		_ = logger.Sync()
	}()

	if len(os.Args) <= 1 {
		zap.L().Warn("Please provide a command: init or run")
		os.Exit(1)
	}
	cmd := os.Args[1]

	// Define CLI argument flags
	purgeCLIArg := flag.Bool("purge", false, "purges the database & all cache files")
	configFilePathCLIArg := flag.String("config", "./config.toml", "config file location (default: ./config.toml)")
	cacheFolderCLIArg := flag.String("cache", "./cache", "cache folder location (default: ./cache)")
	zoneFilesFolderCLIArg := flag.String(
		"zoneFiles", "./zone-files", "zone files folder location (default: ./zone-files)",
	)
	// TODO [HP]: Get rid of queryCLIArg once WebUI is available.
	queryCLIArg := flag.String("query", "", "user query to run after indexing is finished")

	// Parse the CLI arguments
	// Remove the command and leave only the flags in os.Args
	os.Args = append(os.Args[:1], os.Args[2:]...)

	flag.Parse()

	// Override CLI arguments as needed by Environmental Variables
	overridePathArgWithEnvVar(configFilePathCLIArg, "ZFSE_CONF_FILE_PATH")
	overridePathArgWithEnvVar(cacheFolderCLIArg, "ZFSE_CACHE_DIR")
	overridePathArgWithEnvVar(zoneFilesFolderCLIArg, "ZFSE_ZONE_FILES_DIR")

	// Finally set default paths
	path_manager.SetConfigFilePath(*configFilePathCLIArg)
	path_manager.SetCacheFolderPath(*cacheFolderCLIArg)
	path_manager.SetZoneFilesFolderPath(*zoneFilesFolderCLIArg)

	// Check if 'config' command is executed
	if cmd == "init" {
		initializeFilesAndFolders()
	} else if cmd == "run" {
		if *purgeCLIArg {
			purgeCache()
		}

		// Create the application
		conf := config.NewApplicationConfig()
		a := app.NewApplication(conf)

		a.Run(*queryCLIArg)

	} else {
		zap.L().Warn(fmt.Sprintf("Unknown command: %s\n", os.Args[1]))
		os.Exit(1)
	}
}

func overridePathArgWithEnvVar(variable *string, envVariable string) {
	envVar, bEnvVarExists := os.LookupEnv(envVariable)
	if bEnvVarExists {
		path, err := filepath.Abs(envVar)
		if err != nil {
			zap.L().Fatal(
				fmt.Sprintf("Unable to convert path: %v to absolute path.", path),
				zap.String("err", err.Error()),
			)
		}
		// Confirm that this is a valid path
		if !helper.IsValidPath(path) {
			zap.L().Fatal("Path is not valid.", zap.String("path", path))
		}

		zap.L().Info(fmt.Sprintf("Using %v environment variable to override configuration.", envVariable))
		*variable = envVar
	}
}

func initializeFilesAndFolders() {
	// Create the default config
	defaultConfigFilePath := path_manager.GetConfigFilePath()
	defaultConfigFile, err := os.OpenFile(defaultConfigFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		zap.L().Fatal("Unable to create default configuration file.", zap.String("err", err.Error()))
	}
	defer defaultConfigFile.Close()

	defaultConfig := embedding.GetDefaultConfigFile()
	if _, err := defaultConfigFile.WriteString(string(defaultConfig)); err != nil {
		zap.L().Fatal("Unable to write default configuration file.", zap.String("err", err.Error()))
	}

	// Create the cache folder
	err = helper.CreateFolder(path_manager.GetCacheFolderPath())
	if err != nil {
		zap.L().Fatal("Unable to create cache folder.", zap.String("err", err.Error()))
	}

	// Create zone files folder
	err = helper.CreateFolder(path_manager.GetZoneFilesFolderPath())
	if err != nil {
		zap.L().Fatal("Unable to create zone files folder.", zap.String("err", err.Error()))
	}

	// Create the example zone file
	zoneFilesPath := path_manager.GetZoneFilesFolderPath()
	exampleZoneFilePath := filepath.Join(zoneFilesPath, "example_zone_file.txt")
	exampleZoneFile, err := os.OpenFile(exampleZoneFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		zap.L().Fatal("Unable to create example zone file.", zap.String("err", err.Error()))
	}
	defer exampleZoneFile.Close()

	exampleZoneFileText := embedding.GetExampleZoneFile()
	if _, err := exampleZoneFile.WriteString(string(exampleZoneFileText)); err != nil {
		zap.L().Fatal("Unable to write example zone file.", zap.String("err", err.Error()))
	}

	// Create the plugin folder
	err = helper.CreateFolder(path_manager.GetPluginFolderPath())
	if err != nil {
		zap.L().Fatal("Unable to create plugin folder.", zap.String("err", err.Error()))
	}
}

// TODO [MP]: This method is intended as a development helper but we should actually
// allow per zone file cache deletion.
func purgeCache() {
	zap.L().Warn("Purging application cache folders.")

	// Delete cache folders
	err := helper.DeleteFolder(path_manager.GetCacheFolderPath())
	if err != nil {
		zap.L().Fatal("Unable to purge cache folder.", zap.String("err", err.Error()))
	}

	zap.L().Info("Application cache folders successfully deleted.")
}
