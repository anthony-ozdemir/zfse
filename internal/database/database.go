package database

import (
	"encoding/json"

	badger "github.com/dgraph-io/badger/v3"
	"go.uber.org/zap"
)

const dbSchemeVersion = "1.0"

type Database struct {
	db *badger.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	opts := badger.DefaultOptions(dbPath)
	// Design Note: We need to minimize the badgerDB size as much as possible
	// both on disk & memory.
	// TODO [LP]: Check if these options are really suitable.
	opts.VLogPercentile = 0.5
	opts.MemTableSize = 1024 * 1024 * 4
	opts.ValueLogFileSize = 1024 * 1024 * 2
	opts.ValueLogMaxEntries = 1000
	opts.ValueThreshold = int64(float64(opts.ValueLogFileSize) * opts.VLogPercentile / 2)
	opts.CompactL0OnClose = true
	opts.Logger = nil

	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	// Check d scheme version
	d := Database{db: badgerDB}
	dbSchemeVersion, err := d.getString("db_scheme_version")
	if err != nil {
		// Scheme version doesn't exist yet. It means we need to initialize the database.
		zap.L().Info("Initializing the database for the first time.")
		d.initialize()
	} else {
		// Let's check if d scheme has changed
		if dbSchemeVersion != "1.0" {
			zap.L().Warn("Database scheme version has changed. Resetting the database.")
			d.initialize()
		}
	}

	return &d, nil
}

func (d *Database) Purge() {
	err := d.db.DropAll()
	if err != nil {
		zap.L().Fatal("Unable to purge database.", zap.String("err", err.Error()))
	}
}

func (d *Database) initialize() {
	errSetString := d.setString("db_scheme_version", dbSchemeVersion)
	if errSetString != nil {
		zap.L().Fatal("Unable to set key in database.", zap.String("err", errSetString.Error()))
	}
}

type PreCrawlFilterTaskState struct {
	BIsFinished bool `json:"b_is_finished"`
	LineIndex   int  `json:"line_index"`
}

func (d *Database) SavePreCrawlFilterTaskState(zoneName string, state PreCrawlFilterTaskState) {
	// Convert the struct to JSON
	jsonBytes, errMarshal := json.Marshal(state)
	if errMarshal != nil {
		zap.L().Fatal("Error marshaling JSON.", zap.String("err", errMarshal.Error()))
	}

	// Convert the JSON byte slice to a string
	jsonString := string(jsonBytes)

	errSave := d.setString(zoneName+"_pre_crawl_filter_task_state_json", jsonString)
	if errSave != nil {
		zap.L().Fatal("Unable to save task state.")
	}
}

func (d *Database) GetPreCrawlFilterTaskState(zoneName string) PreCrawlFilterTaskState {
	taskStateString, errGet := d.getString(zoneName + "_pre_crawl_filter_task_state_json")
	if errGet != nil {
		zap.L().Info("Unable to get task state for " + zoneName + ". Creating new task state instead.")
		return PreCrawlFilterTaskState{
			BIsFinished: false,
			LineIndex:   0,
		}
	}

	taskState := PreCrawlFilterTaskState{}
	errUnmarshal := json.Unmarshal([]byte(taskStateString), &taskState)
	if errUnmarshal != nil {
		zap.L().Fatal("Error unmarshaling JSON", zap.String("err", errUnmarshal.Error()))
	}

	return taskState
}

type PostCrawlFilterTaskState struct {
	BIsFinished bool `json:"b_is_finished"`
	LineIndex   int  `json:"line_index"`
}

func (d *Database) SavePostCrawlFilterTaskState(zoneName string, state PostCrawlFilterTaskState) {
	// Convert the struct to JSON
	jsonBytes, errMarshal := json.Marshal(state)
	if errMarshal != nil {
		zap.L().Fatal("Error marshaling JSON", zap.String("err", errMarshal.Error()))
	}

	// Convert the JSON byte slice to a string
	jsonString := string(jsonBytes)

	errSave := d.setString(zoneName+"_post_crawl_filter_task_state_json", jsonString)
	if errSave != nil {
		zap.L().Fatal("Unable to save task state.")
	}
}

func (d *Database) GetPostCrawlFilterTaskState(zoneName string) PostCrawlFilterTaskState {
	taskStateString, errGet := d.getString(zoneName + "_post_crawl_filter_task_state_json")
	if errGet != nil {
		zap.L().Info("Unable to get task state for " + zoneName + ". Creating new task state instead.")
		return PostCrawlFilterTaskState{
			BIsFinished: false,
			LineIndex:   0,
		}
	}

	taskState := PostCrawlFilterTaskState{}
	errUnmarshal := json.Unmarshal([]byte(taskStateString), &taskState)
	if errUnmarshal != nil {
		zap.L().Fatal("Error unmarshaling JSON", zap.String("err", errUnmarshal.Error()))
	}

	return taskState
}

type IndexerTaskState struct {
	BIsFinished bool `json:"b_is_finished"`
	LineIndex   int  `json:"line_index"`
}

func (d *Database) SaveIndexerTaskState(zoneName string, state IndexerTaskState) {
	// Convert the struct to JSON
	jsonBytes, errMarshal := json.Marshal(state)
	if errMarshal != nil {
		zap.L().Fatal("Error marshaling JSON", zap.String("err", errMarshal.Error()))
	}

	// Convert the JSON byte slice to a string
	jsonString := string(jsonBytes)

	errSave := d.setString(zoneName+"_"+"indexer_task_state_json", jsonString)
	if errSave != nil {
		zap.L().Fatal("Unable to save task state.")
	}
}

func (d *Database) GetIndexerTaskState(zoneName string) IndexerTaskState {
	taskStateString, errGet := d.getString(zoneName + "_" + "indexer_task_state_json")
	if errGet != nil {
		zap.L().Info("Unable to get indexer task state. Creating new task state instead.")
		indexerTaskState := IndexerTaskState{
			BIsFinished: false,
			LineIndex:   0,
		}
		return indexerTaskState
	}

	taskState := IndexerTaskState{}
	errUnmarshal := json.Unmarshal([]byte(taskStateString), &taskState)
	if errUnmarshal != nil {
		zap.L().Fatal("Error unmarshaling JSON", zap.String("err", errUnmarshal.Error()))
	}

	return taskState
}

func (d *Database) getString(key string) (string, error) {
	var value string
	err := d.db.View(
		func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}
			val, _ := item.ValueCopy(nil)
			value = string(val)
			return nil
		},
	)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *Database) setString(key, value string) error {
	err := d.db.Update(
		func(txn *badger.Txn) error {
			return txn.Set([]byte(key), []byte(value))
		},
	)
	return err
}

func (d *Database) Close() error {
	return d.db.Close()
}
