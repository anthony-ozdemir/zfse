package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/anthony-ozdemir/zfse/internal/config"
	"github.com/anthony-ozdemir/zfse/internal/crawler"
	"github.com/anthony-ozdemir/zfse/internal/database"
	"github.com/anthony-ozdemir/zfse/internal/enum"
	"github.com/anthony-ozdemir/zfse/internal/filebuf"
	"github.com/anthony-ozdemir/zfse/internal/interfaces"
	"github.com/anthony-ozdemir/zfse/internal/metrics_manager"
	"github.com/anthony-ozdemir/zfse/internal/path_manager"
	"github.com/anthony-ozdemir/zfse/internal/task_handlers/indexers"
	"github.com/anthony-ozdemir/zfse/internal/task_handlers/post_crawl_filters"
	"github.com/anthony-ozdemir/zfse/internal/task_handlers/pre_crawl_filters"
	"github.com/anthony-ozdemir/zfse/internal/task_handlers/rankers"
)

type Application struct {
	// Config
	config config.ApplicationConfig

	// Managers
	db                      *database.Database
	metricsManager          *metrics_manager.MetricsManager
	crawler                 *crawler.Crawler
	httpServer              *http.Server
	applicationStateManager *ApplicationStateManager

	// Zone File Registry
	zoneFileRegistry map[string]string

	// Task Handler Registry
	preCrawlFilterRegistry  map[string]interfaces.PreConnectionFilter
	postCrawlFilterRegistry map[string]interfaces.PostConnectionFilter
	indexerRegistry         map[string]interfaces.Indexer
	rankerRegistry          map[string]interfaces.Ranker

	// Task Handler Array
	preCrawlFilterArray  []*interfaces.PreConnectionFilter
	postCrawlFilterArray []*interfaces.PostConnectionFilter
	indexer              *interfaces.Indexer
	rankerArray          []*interfaces.Ranker

	// Normalized Ranker Weight Map
	rankerWeightMap map[string]float64
}

func NewApplication(applicationConfig config.ApplicationConfig) *Application {
	a := Application{
		// Initialize and allocate all built-in containers
		// Zone File Registry
		zoneFileRegistry: make(map[string]string),
		// Task Handler Registry
		preCrawlFilterRegistry:  make(map[string]interfaces.PreConnectionFilter),
		postCrawlFilterRegistry: make(map[string]interfaces.PostConnectionFilter),
		indexerRegistry:         make(map[string]interfaces.Indexer),
		rankerRegistry:          make(map[string]interfaces.Ranker),
		// Task Handler Array
		preCrawlFilterArray:  make([]*interfaces.PreConnectionFilter, 0),
		postCrawlFilterArray: make([]*interfaces.PostConnectionFilter, 0),
		indexer:              nil,
		rankerArray:          make([]*interfaces.Ranker, 0),
		// Normalized Ranker Weight Map
		rankerWeightMap: make(map[string]float64),
	}

	// Set Config
	a.config = applicationConfig

	// Set State Manager
	a.applicationStateManager = NewApplicationStateManager()

	// Register built-in Task Handlers
	// TODO [MP]: We need to create a registry manager system so that plugins can also register Task Handlers
	a.registerPreCrawlFilters()
	a.registerPostCrawlFilters()
	a.registerIndexers()
	a.registerRankers()

	// Initialize Task Handlers
	a.initializeTaskHandlers()

	// Setup metrics
	a.metricsManager = metrics_manager.New()

	// Prepare zoneFileRegistry
	zoneFilesFolderPath := path_manager.GetZoneFilesFolderPath()
	zoneFiles, err := os.ReadDir(zoneFilesFolderPath)
	if err != nil {
		zap.L().Fatal(
			"Unable to read zone zoneFiles directory.",
			zap.String("path", zoneFilesFolderPath),
			zap.String("err", err.Error()),
		)
	}

	for _, zoneFile := range zoneFiles {
		filePath := filepath.Join(zoneFilesFolderPath, zoneFile.Name())
		if !zoneFile.IsDir() {
			// Design Note: Remove extension from zoneFile name
			zoneFileName := strings.TrimSuffix(zoneFile.Name(), filepath.Ext(zoneFile.Name()))

			a.zoneFileRegistry[zoneFileName] = filePath
		}
	}

	if len(a.zoneFileRegistry) <= 0 {
		zap.L().Fatal(
			fmt.Sprintf(
				"No zone files found. Please ensure that zone files are correctly "+
					"located under %v", path_manager.GetZoneFilesFolderPath(),
			),
		)
	}

	// Setup HTTP-Server
	a.httpServer = &http.Server{
		Handler:        a.getRouter(),
		Addr:           a.config.GeneralOptions.ListenAddr + ":" + a.config.GeneralOptions.ListenPort,
		ReadTimeout:    time.Duration(a.config.GeneralOptions.ServerTimeoutInSeconds) * time.Second,
		WriteTimeout:   time.Duration(a.config.GeneralOptions.ServerTimeoutInSeconds) * time.Second,
		IdleTimeout:    time.Duration(a.config.GeneralOptions.ServerTimeoutInSeconds) * time.Second,
		MaxHeaderBytes: 1024, // Max header of 1kb // TODO [LP]: We should make this an option.
		ErrorLog:       zap.NewStdLog(zap.L()),
	}

	// Setup crawler
	crawlerOpts := crawler.CrawlerOptions{
		TimeOutInSeconds:        a.config.GeneralOptions.RequestTimeoutInSeconds,
		MinContentLengthInBytes: a.config.GeneralOptions.MinContentLengthInBytes,
		MaxContentLengthInBytes: a.config.GeneralOptions.MaxContentLengthInBytes,
		ContentReadLimitInBytes: a.config.GeneralOptions.ContentReadLimitInBytes,
	}
	a.crawler = crawler.NewCrawler(crawlerOpts)

	// Setup database
	dbPath := filepath.Join(path_manager.GetCacheFolderPath(), "db")

	db, errDB := database.NewDatabase(dbPath)
	if errDB != nil {
		zap.L().Fatal(
			"Unable to create database.",
			zap.String("path", dbPath),
			zap.String("err", errDB.Error()),
		)
	}

	a.db = db

	return &a
}

// Functions for registering Task Handlers.
func (a *Application) registerPreCrawlFilters() {
	zap.L().Info("Registering built-in Pre-Crawl Filters")
	a.preCrawlFilterRegistry["builtin.unique_domain"] = &pre_crawl_filters.UniqueDomainFilter{}
	a.preCrawlFilterRegistry["builtin.discard_high_entropy"] = &pre_crawl_filters.EntropyFilter{}
	a.preCrawlFilterRegistry["builtin.length_filter"] = &pre_crawl_filters.LengthFilter{}
}

func (a *Application) registerPostCrawlFilters() {
	zap.L().Info("Registering built-in Post-Crawl Filters")
	a.postCrawlFilterRegistry["builtin.description_filter"] = &post_crawl_filters.DescriptionFilter{}
}

func (a *Application) registerIndexers() {
	zap.L().Info("Registering built-in Indexers")
	a.indexerRegistry["builtin.random_indexer"] = &indexers.RandomIndexer{}
	a.indexerRegistry["builtin.basic_indexer"] = &indexers.BasicIndexer{}
}

func (a *Application) registerRankers() {
	zap.L().Info("Registering built-in Rankers")
	a.rankerRegistry["builtin.random_ranker"] = &rankers.RandomRanker{}
	a.rankerRegistry["builtin.indexer_ranker"] = &rankers.IndexerRanker{}
}

// Initializes Task Handlers (PreCrawlFilter, PostCrawlFilter, Indexer, Ranker).
func (a *Application) initializeTaskHandlers() {
	zap.L().Info("Initializing Task Handlers")
	// PreCrawlFilters
	for _, preCrawlFilterOptions := range a.config.PreCrawlFilterOptions {
		preCrawlFilterType := preCrawlFilterOptions.Type
		preCrawlFilter, ok := a.preCrawlFilterRegistry[preCrawlFilterType]
		if !ok {
			zap.L().Fatal(
				"Pre-crawl Filter not found in registry.",
				zap.String("pre_crawl_filter_type", preCrawlFilterType),
			)
		}
		err := preCrawlFilter.Initialize(preCrawlFilterOptions)
		if err != nil {
			zap.L().Fatal(
				"Unable to initialize Pre-crawl Filter.",
				zap.String("pre_crawl_filter_type", preCrawlFilterType),
				zap.String("err", err.Error()),
			)
		}
		a.preCrawlFilterArray = append(a.preCrawlFilterArray, &preCrawlFilter)
	}

	// PostCrawlFilters
	for _, postCrawlFilterOptions := range a.config.PostCrawlFilterOptions {
		postCrawlFilterType := postCrawlFilterOptions.Type
		postCrawlFilter, ok := a.postCrawlFilterRegistry[postCrawlFilterType]
		if !ok {
			zap.L().Fatal(
				"Post-crawl filter not found in registry.",
				zap.String("post_crawl_filter_type", postCrawlFilterType),
			)
		}
		err := postCrawlFilter.Initialize(postCrawlFilterOptions)
		if err != nil {
			zap.L().Fatal(
				"Unable to initialize Task Handler.",
				zap.String("post_crawl_filter_type", postCrawlFilterType),
				zap.String("err", err.Error()),
			)
		}
		a.postCrawlFilterArray = append(a.postCrawlFilterArray, &postCrawlFilter)
	}

	// Indexer
	{
		indexerType := a.config.IndexerOption.Type
		indexer, ok := a.indexerRegistry[indexerType]
		if !ok {
			zap.L().Fatal(
				"Indexer not found in registry.",
				zap.String("indexer_type", indexerType),
			)
		}
		err := indexer.Initialize(
			a.config.IndexerOption, path_manager.GetIndexerDatabaseFilePath(indexerType),
			a.config.GeneralOptions.IndexerOutputLimit,
		)
		if err != nil {
			zap.L().Fatal(
				"Unable to initialize Indexer.",
				zap.String("indexer_type", indexerType),
				zap.String("err", err.Error()),
			)
		}
		a.indexer = &indexer
	}

	// Rankers
	totalRankerWeights := 0.0
	for _, rankerOptions := range a.config.RankerOptions {
		rankerType := rankerOptions.Type
		ranker, ok := a.rankerRegistry[rankerType]
		if !ok {
			zap.L().Fatal(
				"Ranker not found in registry.",
				zap.String("ranker_type", rankerType),
			)
		}
		err := ranker.Initialize(rankerOptions)
		if err != nil {
			zap.L().Fatal(
				"Unable to initialize Ranker.",
				zap.String("ranker_type", rankerType),
				zap.String("err", err.Error()),
			)
		}
		a.rankerArray = append(a.rankerArray, &ranker)
		rankerWeight := rankerOptions.FloatOptions["weight"]
		totalRankerWeights += rankerWeight
	}

	// Let's normalize ranker weights
	for _, ranker := range a.rankerArray {
		rankerType := (*ranker).GetType()
		foundRankerOption := config.TaskHandlerOptions{}
		for _, taskHandlerConfig := range a.config.RankerOptions {
			if taskHandlerConfig.Type == rankerType {
				foundRankerOption = taskHandlerConfig
				break
			}
		}
		rankerWeight := foundRankerOption.FloatOptions["weight"]
		normalizedRankerWeight := rankerWeight / totalRankerWeights
		a.rankerWeightMap[rankerType] = normalizedRankerWeight
	}
}

func (a *Application) Rank(userQuery string) {
	a.applicationStateManager.OnRankingStarted()
	zap.L().Info("Starting query.", zap.String("user_query", userQuery))
	// Query
	indexerOutput := a.queryIndexer(userQuery)

	// Ready to rank at this stage.
	rankerOutput := a.runRankers(indexerOutput, userQuery)

	// Let's cache the ranker output
	rankerOutputBufferOpts := filebuf.FileOutputBufferOptions{
		BulkOutputLimit: a.config.GeneralOptions.FileBulkOutputQty,
		FilePath:        path_manager.GetRankingFilePath(""),
	}
	rankerOutputBuffer := filebuf.NewFileOutputBuffer(rankerOutputBufferOpts)

	for _, output := range rankerOutput {
		jsonString, err := output.ToJSONString()
		if err != nil {
			zap.L().Fatal("Unable to unmarshall JSON", zap.String("err", err.Error()))
		}
		rankerOutputBuffer.AppendToFile(jsonString)
	}
	rankerOutputBuffer.Flush()

	// TODO [HP]: Delete once WebUI is available
	queryOutput := make([]zap.Field, 0)
	for i, domainProperty := range rankerOutput {
		queryOutput = append(queryOutput, zap.String(strconv.Itoa(i), domainProperty.DomainName))
	}
	zap.L().Info("Ranker results", queryOutput...)

	a.applicationStateManager.OnReadyToSearch()
	zap.L().Info(fmt.Sprintf("Ranking finished. Ready to search."))
}

func (a *Application) outputMetrics(applicationState ApplicationState) {
	zap.L().Info(
		"Metrics",
		zap.String("Task State", applicationState.task.String()),
		zap.Int("total_work_items", applicationState.totalWorkItems),
		zap.Int("remaining_work_items", applicationState.processedWorkItems),
		zap.Int("estimate_remaining_time_inseconds", applicationState.estimateRemainingTimeInSeconds),
	)
}

func (a *Application) Run(userQuery string) {
	// Setup Interrupt Handlers for graceful shut-down & Docker sent SIGTERM.
	interruptChan := make(chan os.Signal, 2)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		interruptCounter := 0
		for {
			select {
			case <-interruptChan:
				interruptCounter++
				if interruptCounter == 1 {
					zap.L().Warn("Received interrupt signal. Starting graceful shutdown...")
					a.applicationStateManager.OnShutdown()
				} else if interruptCounter >= 2 {
					zap.L().Warn("Received second interrupt signal. Quitting immediately...")
					os.Exit(1)
				}
			}
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO [HP]: Implement webUI
		zap.L().Info(fmt.Sprintf("Starting server on http://%v", a.httpServer.Addr))
		err := a.httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			zap.L().Fatal("Server launch failed.", zap.String("err", err.Error()))
		}
	}()

	ticker := time.NewTicker(time.Duration(a.config.GeneralOptions.MetricOutputPerSeconds) * time.Second)
	defer ticker.Stop()

	var once sync.Once // TODO [HP]: Get rid of this variable once WebUI is available
	for {
		// Check if we should start graceful shutdown
		currentState := a.applicationStateManager.GetApplicationState()
		if currentState.task == enum.Shutdown {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := a.httpServer.Shutdown(ctx); err != nil {
				zap.L().Fatal("Unable to shut down the server gracefully.", zap.String("err", err.Error()))
			}
			break
		}

		select {
		case <-ticker.C:
			// Output Metrics
			a.outputMetrics(currentState)
		default:
			// TODO [MP]: A bit hacky code, perhaps we can just allow these to call & chain themselves
			if currentState.task == enum.Initializing {
				// Let's start pre-crawl filtering
				a.applicationStateManager.OnPreCrawlFiltersStarted()
				wg.Add(1)
				go func() {
					a.runPreCrawlFilters(&wg)
				}()
			} else if currentState.task == enum.FinishedPreCrawlFilters {
				// Let's start post-crawl filtering
				a.applicationStateManager.OnPostCrawlFiltersStarted()
				wg.Add(1)
				go func() {
					a.runPostCrawlFilters(&wg)
				}()
			} else if currentState.task == enum.FinishedPostCrawlFilters {
				// Let's start indexing
				a.applicationStateManager.OnIndexingStarted()
				wg.Add(1)
				go func() {
					a.runIndexer(&wg)
				}()
			} else if currentState.task == enum.ReadyToSearch {
				// TODO [HP]: We need to get rid of code here once WebUI is available.
				once.Do(
					func() {
						if userQuery != "" {
							a.Rank(userQuery)
						}
					},
				)
			}

			time.Sleep(1 * time.Second)
		}
	}

	wg.Wait() // Wait for all goroutines to finish

	errClose := a.db.Close()
	if errClose != nil {
		zap.L().Fatal("Unable to close the db.", zap.String("err", errClose.Error()))
	}
	zap.L().Info("Application shut-down.")

}
