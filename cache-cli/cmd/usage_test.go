package cmd

import (
	"fmt"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	assert "github.com/stretchr/testify/assert"
)

func Test__Usage(t *testing.T) {
	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s empty cache", backend), func(t *testing.T) {
			storage.Clear()

			capturer := utils.CreateOutputCapturer()
			RunUsage(usageCmd, []string{})
			output := capturer.Done()

			switch backend {
			case "s3":
				assert.Contains(t, output, "FREE SPACE: (unlimited)")
				assert.Contains(t, output, "USED SPACE: 0B")
			case "sftp":
				assert.Contains(t, output, "FREE SPACE: 9.0G")
				assert.Contains(t, output, "USED SPACE: 0B")
			}
		})
	})
}
