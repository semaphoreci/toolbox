package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/logging"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	log "github.com/sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

func Test__Clear(t *testing.T) {
	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s no keys", backend), func(*testing.T) {
			err := storage.Clear()
			assert.Nil(t, err)

			RunClear(clearCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Deleted all caches.")
		})

		t.Run(fmt.Sprintf("%s with keys", backend), func(*testing.T) {
			err := storage.Clear()
			assert.Nil(t, err)

			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			storage.Store("abc001", tempFile.Name())

			RunClear(hasKeyCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Deleted all caches.")
		})
	})
}
