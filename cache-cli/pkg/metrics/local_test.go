package metrics

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Publish(t *testing.T) {
	os.Setenv("SEMAPHORE_EXECUTION_ENVIRONMENT", "hosted")
	metricsManager, err := NewLocalMetricsBackend()
	assert.Nil(t, err)

	t.Run("valid cache metrics", func(t *testing.T) {
		err = metricsManager.PublishBatch([]Metric{
			{Name: "cache_download_size", Value: "1000"},
			{Name: "cache_download_time", Value: "30"},
			{Name: "cache_user", Value: "tester"},
			{Name: "cache_server", Value: "0.0.0.0"},
		})

		assert.Nil(t, err)

		bytes, err := ioutil.ReadFile(metricsManager.CacheMetricsPath)
		assert.Nil(t, err)

		assert.Contains(t, string(bytes), "cache_download_size 1000")
		assert.Contains(t, string(bytes), "cache_download_time 30")
		assert.Contains(t, string(bytes), "cache_user tester")
		assert.Contains(t, string(bytes), "cache_server 0.0.0.0")

		_, err = os.Stat(metricsManager.ToolboxMetricsPath)
		assert.NotNil(t, err)

		os.Remove(metricsManager.CacheMetricsPath)
	})

	t.Run("valid toolbox metrics", func(t *testing.T) {
		err = metricsManager.PublishBatch([]Metric{
			{Name: "cache_total_rate", Value: "1"},
			{Name: "cache_corruption_rate", Value: "1"},
		})

		assert.Nil(t, err)

		bytes, err := ioutil.ReadFile(metricsManager.ToolboxMetricsPath)
		assert.Nil(t, err)

		assert.Contains(t, string(bytes), "cache_total_rate 1")
		assert.Contains(t, string(bytes), "cache_corruption_rate 1")

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
		os.Setenv("SEMAPHORE_EXECUTION_ENVIRONMENT", "self-hosted")

		err = metricsManager.PublishBatch([]Metric{
			{Name: "cache_download_size", Value: "1000"},
			{Name: "cache_download_time", Value: "30"},
			{Name: "cache_user", Value: "tester"},
			{Name: "cache_server", Value: "0.0.0.0"},
			{Name: "cache_total_rate", Value: "1"},
			{Name: "cache_corruption_rate", Value: "1"},
			{Name: "some-invalid-metric-name", Value: "invalid"},
		})

		assert.Nil(t, err)

		_, err = os.Stat(metricsManager.CacheMetricsPath)
		assert.NotNil(t, err)

		_, err = os.Stat(metricsManager.ToolboxMetricsPath)
		assert.NotNil(t, err)
	})
}
