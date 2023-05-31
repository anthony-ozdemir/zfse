package config

import (
	"fmt"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"

	"github.com/anthony-ozdemir/zfse/internal/path_manager"
)

const currentConfigVersion = "1.0"

type ApplicationConfigVersionOnly struct {
	GeneralOptions struct {
		Version string `toml:"version"`
	} `toml:"General"`
}

type ApplicationConfig struct {
	GeneralOptions         GeneralOptions       `toml:"General"`
	PreCrawlFilterOptions  []TaskHandlerOptions `toml:"PreCrawlFilters"`
	PostCrawlFilterOptions []TaskHandlerOptions `toml:"PostCrawlFilters"`
	IndexerOption          TaskHandlerOptions   `toml:"Indexer"`
	RankerOptions          []TaskHandlerOptions `toml:"Rankers"`
}

type GeneralOptions struct {
	ListenAddr             string `toml:"listen_addr"`
	ListenPort             string `toml:"listen_port"`
	ServerTimeoutInSeconds int    `toml:"server_timeout_in_seconds"`

	FileBulkOutputQty int64 `toml:"file_bulk_output_qty"`

	NumThreadHint int `toml:"num_thread_hint"`

	LogFile                 bool   `toml:"log_file"`
	LogConsole              bool   `toml:"log_console"`
	MetricOutputPerSeconds  int    `toml:"metric_output_per_seconds"`
	ConnectionProtocol      string `toml:"connection_protocol"`
	RequestTimeoutInSeconds int    `toml:"request_timeout_in_seconds"`
	MinContentLengthInBytes int64  `toml:"min_content_length_in_bytes"`
	MaxContentLengthInBytes int64  `toml:"max_content_length_in_bytes"`
	ContentReadLimitInBytes int64  `toml:"content_read_limit_in_bytes"`
	ConcurrentConnections   int    `toml:"concurrent_connections"`

	IndexerOutputLimit int64 `toml:"indexer_output_limit"`
}

type TaskHandlerOptions struct {
	Type          string
	StringOptions map[string]string
	IntOptions    map[string]int64
	FloatOptions  map[string]float64
	BoolOptions   map[string]bool
}

// Design Note: Custom Unmarshaller for FilterTaskType Handlers
// This is needed as we will eventually add plugins.
func (f *TaskHandlerOptions) UnmarshalTOML(data interface{}) error {
	rawData, _ := data.(map[string]interface{})

	// Initialize maps
	if f.StringOptions == nil {
		f.StringOptions = make(map[string]string)
	}

	if f.IntOptions == nil {
		f.IntOptions = make(map[string]int64)
	}

	if f.FloatOptions == nil {
		f.FloatOptions = make(map[string]float64)
	}

	if f.BoolOptions == nil {
		f.BoolOptions = make(map[string]bool)
	}

	// Every FilterTaskType Handler needs to have a type field
	f.Type = rawData["type"].(string)

	for key, value := range rawData {
		switch reflect.TypeOf(value).Kind() {
		case reflect.String:
			f.StringOptions[key] = value.(string)
		case reflect.Float64:
			f.FloatOptions[key] = value.(float64)
		case reflect.Int64:
			f.IntOptions[key] = value.(int64)
		case reflect.Bool:
			f.BoolOptions[key] = value.(bool)
		default:
			// Ignore unknown value types
		}
	}

	return nil
}

func NewApplicationConfig() ApplicationConfig {
	zap.L().Info("Started parsing configuration.")

	configBytes, err := os.ReadFile(path_manager.GetConfigFilePath())
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	// Let's check the version of the config file
	configVersionOnly := ApplicationConfigVersionOnly{}
	if _, err := toml.Decode(string(configBytes), &configVersionOnly); err != nil {
		zap.L().Fatal(err.Error())
	}

	if configVersionOnly.GeneralOptions.Version != currentConfigVersion {
		zap.L().Fatal(
			fmt.Sprintf(
				"Existing config.toml version does not match the current version of %v.\n"+
					" Please either manually edit the existing configuration to the match latest configuration"+
					" format or delete it.\n Exiting application.",
				currentConfigVersion,
			),
		)
	}

	config := ApplicationConfig{}
	if _, err := toml.Decode(string(configBytes), &config); err != nil {
		zap.L().Fatal(err.Error())
	}

	// Sanity Check
	// Every Ranker needs a float type "weight" option to be specified
	for _, rankerOption := range config.RankerOptions {
		_, bHasWeight := rankerOption.FloatOptions["weight"]
		if !bHasWeight {
			zap.L().Fatal(
				"Invalid Ranker configuration. \"weight\" option of float type is not specified.",
				zap.String("ranker_type", rankerOption.Type),
			)
		}
	}

	zap.L().Info("Finished parsing configuration.")
	return config
}
