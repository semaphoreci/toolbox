package cmd

import (
	"fmt"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/logging"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	log "github.com/sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

func Test__Usage(t *testing.T) {
	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s empty cache", backend), func(t *testing.T) {
			storage.Clear()

			RunUsage(usageCmd, []string{})
			output := readOutputFromFile(t)

			switch backend {
			case "s3":
				assert.Contains(t, output, "FREE SPACE: (unlimited)")
				assert.Contains(t, output, "USED SPACE: 0.0")
			case "sftp":
				assert.Contains(t, output, "FREE SPACE: 9.0G")
				assert.Contains(t, output, "USED SPACE: 0.0")
			}
		})
	})
}
