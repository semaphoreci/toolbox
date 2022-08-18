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

func Test__Delete(t *testing.T) {
	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s key is missing", backend), func(*testing.T) {
			RunDelete(deleteCmd, []string{"this-key-does-not-exist"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Key 'this-key-does-not-exist' doesn't exist in the cache store.")
		})

		t.Run(fmt.Sprintf("%s key is present", backend), func(*testing.T) {
			storage.Clear()
			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			storage.Store("abc001", tempFile.Name())

			RunDelete(deleteCmd, []string{"abc001"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Key 'abc001' is deleted.")
		})

		t.Run(fmt.Sprintf("%s normalizes key", backend), func(*testing.T) {
			storage.Clear()
			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			RunStore(storeCmd, []string{"abc/00/33", tempFile.Name()})

			RunDelete(deleteCmd, []string{"abc/00/33"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Key 'abc/00/33' is normalized to 'abc-00-33'")
			assert.Contains(t, output, "Key 'abc-00-33' is deleted.")
		})
	})
}
