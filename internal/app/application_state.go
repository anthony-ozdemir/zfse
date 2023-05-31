package app

import (
	"sync"
	"time"

	"github.com/anthony-ozdemir/zfse/internal/enum"
)

type ApplicationState struct {
	task                           enum.ApplicationTask
	taskStartTime                  time.Time
	totalWorkItems                 int
	processedWorkItems             int
	remainingWorkItems             int
	estimateRemainingTimeInSeconds int
	errorDetails                   string
}

type ApplicationStateManager struct {
	mutex            sync.Mutex
	applicationState ApplicationState
}

func NewApplicationStateManager() *ApplicationStateManager {
	return &ApplicationStateManager{
		applicationState: ApplicationState{
			task:                           enum.Initializing,
			totalWorkItems:                 0,
			processedWorkItems:             0,
			estimateRemainingTimeInSeconds: 0,
			errorDetails:                   "",
		},
	}
}

func (a *ApplicationStateManager) GetApplicationState() ApplicationState {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.applicationState
}

func (a *ApplicationStateManager) SetTotalWorkItems(totalWorkItems int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.applicationState.totalWorkItems = totalWorkItems
}

func (a *ApplicationStateManager) SetProcessedWorkItems(processedWorkItems int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.applicationState.processedWorkItems = processedWorkItems
	a.applicationState.remainingWorkItems = a.applicationState.totalWorkItems - processedWorkItems

	// Let's also estimate remaining time
	elapsedTime := time.Since(a.applicationState.taskStartTime).Seconds()
	estimateRemainingDurationInSeconds := 0.0
	if processedWorkItems > 0 {
		estimateRemainingDurationInSeconds = (elapsedTime / float64(processedWorkItems)) *
			float64(a.applicationState.remainingWorkItems)
	} else {
		// Just a wild estimate to begin.
		estimateRemainingDurationInSeconds = float64(a.applicationState.remainingWorkItems) * 0.01
	}
	a.applicationState.estimateRemainingTimeInSeconds = int(estimateRemainingDurationInSeconds)
}

// Design Note: We need to create a small state-machine here, where
// application can only move to a limited set of states from an initial state.

func (a *ApplicationStateManager) OnPreCrawlFiltersStarted() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.applicationState.task != enum.Initializing {
		panic("Programming error.")
	}

	a.applicationState.task = enum.RunningPreCrawlFilters
	a.applicationState.taskStartTime = time.Now()
}

func (a *ApplicationStateManager) OnPreCrawlFiltersFinished() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.applicationState.task != enum.RunningPreCrawlFilters {
		panic("Programming error.")
	}

	a.applicationState.task = enum.FinishedPreCrawlFilters
	a.applicationState.totalWorkItems = 0
	a.applicationState.processedWorkItems = 0
	a.applicationState.estimateRemainingTimeInSeconds = 0
}

func (a *ApplicationStateManager) OnPostCrawlFiltersStarted() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.applicationState.task != enum.FinishedPreCrawlFilters {
		panic("Programming error.")
	}

	a.applicationState.task = enum.RunningPostCrawlFilters
	a.applicationState.taskStartTime = time.Now()
}

func (a *ApplicationStateManager) OnPostCrawlFiltersFinished() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.applicationState.task != enum.RunningPostCrawlFilters {
		panic("Programming error.")
	}

	a.applicationState.task = enum.FinishedPostCrawlFilters
	a.applicationState.totalWorkItems = 0
	a.applicationState.processedWorkItems = 0
	a.applicationState.estimateRemainingTimeInSeconds = 0
}

func (a *ApplicationStateManager) OnIndexingStarted() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.applicationState.task != enum.FinishedPostCrawlFilters {
		panic("Programming error.")
	}

	a.applicationState.task = enum.RunningIndexer
	a.applicationState.taskStartTime = time.Now()
}

func (a *ApplicationStateManager) OnReadyToSearch() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.applicationState.task != enum.RunningIndexer && a.applicationState.task != enum.RunningRankers {
		panic("Programming error.")
	}

	a.applicationState.task = enum.ReadyToSearch
	a.applicationState.totalWorkItems = 0
	a.applicationState.processedWorkItems = 0
	a.applicationState.estimateRemainingTimeInSeconds = 0
}

func (a *ApplicationStateManager) OnRankingStarted() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.applicationState.task != enum.ReadyToSearch {
		panic("Programming error.")
	}

	a.applicationState.task = enum.RunningRankers
	a.applicationState.taskStartTime = time.Now()
}

func (a *ApplicationStateManager) OnErrored(errorDetails string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.applicationState.task = enum.Errored
	a.applicationState.errorDetails = errorDetails
}

func (a *ApplicationStateManager) OnShutdown() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.applicationState.task = enum.Shutdown
}
