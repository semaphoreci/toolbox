package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	assert "github.com/stretchr/testify/assert"
)

func Test__UnpackSendsMetricsOnFailure(t *testing.T) {
	os.Setenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED", "true")
	metricsManager, err := metrics.InitMetricsManager(metrics.LocalBackend)
	assert.Nil(t, err)

	tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
	tempFile.WriteString("this is not a proper archive")

	_, err = Unpack(metricsManager, tempFile.Name())
	assert.NotNil(t, err)

	bytes, err := ioutil.ReadFile(fmt.Sprintf("%s/toolbox_metrics", os.TempDir()))
	assert.Nil(t, err)
	assert.Contains(t, string(bytes), fmt.Sprintf("%s 1", metrics.CacheCorruptionRate))

	os.Remove(tempFile.Name())
	os.Remove(fmt.Sprintf("%s/toolbox_metrics", os.TempDir()))
}
