package embedding

import (
	// Blank import is needed as we are storing main.go under "./cmd".
	// TODO [LP]: Find a better solution. This might embed all the files on all binaries.
	_ "embed"
)

//go:embed data/config/config.toml
var defaultConfig []byte

// GetDefaultConfigFile Returns embedded default "config.toml".
func GetDefaultConfigFile() []byte {
	return defaultConfig
}

//go:embed data/zone_files/example_zone_file.txt
var exampleZoneFile []byte

// GetExampleZoneFile Returns embedded "example_zone_file.txt".
func GetExampleZoneFile() []byte {
	return exampleZoneFile
}
