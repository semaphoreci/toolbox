package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/logging"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	log "github.com/sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

func Test__HasKey(t *testing.T) {
	ctx := context.TODO()
	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(ctx, t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s key is missing", backend), func(*testing.T) {
			RunHasKey(hasKeyCmd, []string{"this-key-does-not-exist"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Key 'this-key-does-not-exist' doesn't exist in the cache store.")
		})

		t.Run(fmt.Sprintf("%s key is present", backend), func(*testing.T) {
			storage.Clear(ctx)
			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			storage.Store(ctx, "abc001", tempFile.Name())

			RunHasKey(hasKeyCmd, []string{"abc001"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Key 'abc001' exists in the cache store.")
		})

		t.Run(fmt.Sprintf("%s normalizes key", backend), func(*testing.T) {
			storage.Clear(ctx)
			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			RunStore(NewStoreCommand(), []string{"abc/00/33", tempFile.Name()})

			RunHasKey(hasKeyCmd, []string{"abc/00/33"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Key 'abc/00/33' is normalized to 'abc-00-33'")
			assert.Contains(t, output, "Key 'abc-00-33' exists in the cache store.")
		})
	})
}
