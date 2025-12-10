package metrics

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

func TestLogEventWritesInfluxLine(t *testing.T) {
	os.Setenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED", "true")
	os.Setenv("SEMAPHORE_CACHE_USERNAME", "tester")
	os.Setenv("SEMAPHORE_CACHE_URL", "10.0.0.5:1234")

	metricsManager, err := NewLocalMetricsBackend()
	assert.Nil(t, err)

	event := CacheEvent{
		Command:   CommandStore,
		SizeBytes: 2048,
		Duration:  time.Second,
	}

	err = metricsManager.LogEvent(event)
	assert.Nil(t, err)

	bytes, err := ioutil.ReadFile(metricsManager.ToolboxMetricsPath)
	assert.Nil(t, err)
	assert.Contains(t, string(bytes), "usercache,server=10.0.0.5,user=tester,command=store,corrupt=0 size=2048,duration=1000")

	os.Remove(metricsManager.ToolboxMetricsPath)
}

func TestLogEventMarksCorruption(t *testing.T) {
	os.Setenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED", "true")
	os.Setenv("SEMAPHORE_CACHE_USERNAME", "")
	os.Setenv("SEMAPHORE_CACHE_URL", "")

	metricsManager, err := NewLocalMetricsBackend()
	assert.Nil(t, err)

	event := CacheEvent{
		Command: CommandRestore,
		Corrupt: true,
	}

	err = metricsManager.LogEvent(event)
	assert.Nil(t, err)

	bytes, err := ioutil.ReadFile(metricsManager.ToolboxMetricsPath)
	assert.Nil(t, err)
	assert.Contains(t, string(bytes), "command=restore,corrupt=1")
	assert.Contains(t, string(bytes), "size=0,duration=0")

	os.Remove(metricsManager.ToolboxMetricsPath)
}

func TestLogEventDisabled(t *testing.T) {
	os.Setenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED", "false")
	metricsManager, err := NewLocalMetricsBackend()
	assert.Nil(t, err)

	event := CacheEvent{
		Command:   CommandStore,
		SizeBytes: 100,
	}

	err = metricsManager.LogEvent(event)
	assert.Nil(t, err)

	_, err = os.Stat(metricsManager.ToolboxMetricsPath)
	assert.NotNil(t, err)
}
