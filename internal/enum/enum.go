package enum

//go:generate stringer -type=FilterTaskState
//go:generate stringer -type=IndexerTaskState
//go:generate stringer -type=ApplicationTask

type ApplicationTask int64

const (
	Initializing ApplicationTask = iota
	RunningPreCrawlFilters
	FinishedPreCrawlFilters
	RunningPostCrawlFilters
	FinishedPostCrawlFilters
	RunningIndexer
	ReadyToSearch // Finished Indexing
	RunningRankers
	Errored
	Shutdown
)

type FilterTaskState int64

const (
	PreCrawlFilter FilterTaskState = iota
	PostCrawlFilter
	FinishedFiltering
)

type IndexerTaskState int64

const (
	Indexing IndexerTaskState = iota
	FinishedIndexing
)
