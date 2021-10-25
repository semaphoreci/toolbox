package metrics

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Publish(t *testing.T) {
	os.Setenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED", "true")
	metricsManager, err := NewLocalMetricsBackend()
	assert.Nil(t, err)

	t.Run("valid cache metrics", func(t *testing.T) {
		err = metricsManager.PublishBatch([]Metric{
			{Name: CacheDownloadSize, Value: "1000"},
			{Name: CacheDownloadTime, Value: "30"},
			{Name: CacheUser, Value: "tester"},
			{Name: CacheServer, Value: "0.0.0.0"},
		})

		assert.Nil(t, err)

		bytes, err := ioutil.ReadFile(metricsManager.CacheMetricsPath)
		assert.Nil(t, err)

		assert.Contains(t, string(bytes), fmt.Sprintf("%s 1000", CacheDownloadSize))
		assert.Contains(t, string(bytes), fmt.Sprintf("%s 30", CacheDownloadTime))
		assert.Contains(t, string(bytes), fmt.Sprintf("%s tester", CacheUser))
		assert.Contains(t, string(bytes), fmt.Sprintf("%s 0.0.0.0", CacheServer))

		_, err = os.Stat(metricsManager.ToolboxMetricsPath)
		assert.NotNil(t, err)

		os.Remove(metricsManager.CacheMetricsPath)
	})

	t.Run("valid toolbox metrics", func(t *testing.T) {
		err = metricsManager.PublishBatch([]Metric{
			{Name: CacheTotalRate, Value: "1"},
			{Name: CacheCorruptionRate, Value: "1"},
		})

		assert.Nil(t, err)

		bytes, err := ioutil.ReadFile(metricsManager.ToolboxMetricsPath)
		assert.Nil(t, err)

		assert.Contains(t, string(bytes), fmt.Sprintf("%s 1", CacheTotalRate))
		assert.Contains(t, string(bytes), fmt.Sprintf("%s 1", CacheCorruptionRate))

		_, err = os.Stat(metricsManager.CacheMetricsPath)
		assert.NotNil(t, err)

		os.Remove(metricsManager.ToolboxMetricsPath)
	})

	t.Run("invalid metrics are ignored", func(t *testing.T) {
		err = metricsManager.PublishBatch([]Metric{
			{Name: "some-invalid-metric-name", Value: "invalid"},
		})

		assert.Nil(t, err)

		_, err = os.Stat(metricsManager.CacheMetricsPath)
		assert.NotNil(t, err)

		_, err = os.Stat(metricsManager.ToolboxMetricsPath)
		assert.NotNil(t, err)
	})

	t.Run("ignores metrics if it is not enabled", func(t *testing.T) {
		os.Setenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED", "false")

		err = metricsManager.PublishBatch([]Metric{
			{Name: CacheDownloadSize, Value: "1000"},
			{Name: CacheDownloadTime, Value: "30"},
			{Name: CacheUser, Value: "tester"},
			{Name: CacheServer, Value: "0.0.0.0"},
			{Name: CacheTotalRate, Value: "1"},
			{Name: CacheCorruptionRate, Value: "1"},
			{Name: "some-invalid-metric-name", Value: "invalid"},
		})

		assert.Nil(t, err)

		_, err = os.Stat(metricsManager.CacheMetricsPath)
		assert.NotNil(t, err)

		_, err = os.Stat(metricsManager.ToolboxMetricsPath)
		assert.NotNil(t, err)
	})
}
