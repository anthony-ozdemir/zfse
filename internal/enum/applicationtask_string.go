// Code generated by "stringer -type=ApplicationTask"; DO NOT EDIT.

package enum

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Initializing-0]
	_ = x[RunningPreCrawlFilters-1]
	_ = x[FinishedPreCrawlFilters-2]
	_ = x[RunningPostCrawlFilters-3]
	_ = x[FinishedPostCrawlFilters-4]
	_ = x[RunningIndexer-5]
	_ = x[ReadyToSearch-6]
	_ = x[RunningRankers-7]
	_ = x[Errored-8]
	_ = x[Shutdown-9]
}

const _ApplicationTask_name = "InitializingRunningPreCrawlFiltersFinishedPreCrawlFiltersRunningPostCrawlFiltersFinishedPostCrawlFiltersRunningIndexerReadyToSearchRunningRankersErroredShutdown"

var _ApplicationTask_index = [...]uint8{0, 12, 34, 57, 80, 104, 118, 131, 145, 152, 160}

func (i ApplicationTask) String() string {
	if i < 0 || i >= ApplicationTask(len(_ApplicationTask_index)-1) {
		return "ApplicationTask(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ApplicationTask_name[_ApplicationTask_index[i]:_ApplicationTask_index[i+1]]
}