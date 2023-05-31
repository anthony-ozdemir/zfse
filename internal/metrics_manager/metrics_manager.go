package metrics_manager

import (
	"sync"

	"github.com/rcrowley/go-metrics"
	"go.uber.org/zap"
)

type MetricsManager struct {
	metricsRegistry metrics.Registry
	counterNames    map[string]bool
	mutex           sync.RWMutex
}

func New() *MetricsManager {
	m := MetricsManager{}
	m.metricsRegistry = metrics.NewRegistry()
	m.counterNames = make(map[string]bool)
	return &m
}

func (m *MetricsManager) NewCounter(counterName string) {
	// Write-Lock Mutex
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metricsRegistry.GetOrRegister(counterName, &metrics.StandardCounter{})
	m.counterNames[counterName] = true
}

func (m *MetricsManager) IncCounter(counterName string, amount int64) {
	// Write-Lock Mutex
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metric, ok := m.metricsRegistry.Get(counterName).(*metrics.StandardCounter)
	if !ok {
		zap.L().Fatal("Invalid counter name.", zap.String("counter_name", counterName))
	}
	metric.Inc(amount)
}

func (m *MetricsManager) DecCounter(counterName string, amount int64) {
	// Write-Lock Mutex
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metric, ok := m.metricsRegistry.Get(counterName).(*metrics.StandardCounter)
	if !ok {
		zap.L().Fatal("Invalid counter name.", zap.String("counter_name", counterName))
	}
	metric.Dec(amount)
}

func (m *MetricsManager) GetCounterCount(counterName string) int64 {
	// Read-Lock Mutex
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metric, ok := m.metricsRegistry.Get(counterName).(*metrics.StandardCounter)
	if !ok {
		zap.L().Fatal("Invalid counter name.", zap.String("counter_name", counterName))
	}
	return metric.Count()
}

func (m *MetricsManager) GetAllCounterCounts() map[string]int64 {
	counterCountMap := make(map[string]int64)

	for counterName, _ := range m.counterNames {
		value := m.GetCounterCount(counterName)
		counterCountMap[counterName] = value
	}

	return counterCountMap
}

func (m *MetricsManager) GetAllCounterCountsAsZapFields() []zap.Field {
	// Read-Lock Mutex
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	fields := []zap.Field{}
	for k, v := range m.GetAllCounterCounts() {
		fields = append(fields, zap.Int64(k, v))
	}
	return fields
}

func (m *MetricsManager) Reset() {
	// Write-Lock Mutex
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for counterName, _ := range m.counterNames {
		metric := m.metricsRegistry.Get(counterName).(*metrics.StandardCounter)
		metric.Clear()
	}
}
