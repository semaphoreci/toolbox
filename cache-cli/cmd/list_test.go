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

func Test__List(t *testing.T) {
	ctx := context.TODO()
	listCmd := NewListCommand()
	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(ctx, t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s no keys", backend), func(*testing.T) {
			storage.Clear(ctx)

			RunList(listCmd, []string{""})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Cache is empty.")
		})

		t.Run(fmt.Sprintf("%s with keys", backend), func(*testing.T) {
			storage.Clear(ctx)
			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			storage.Store(ctx, "abc001", tempFile.Name())
			storage.Store(ctx, "abc002", tempFile.Name())
			storage.Store(ctx, "abc003", tempFile.Name())

			RunList(listCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "abc001")
			assert.Contains(t, output, "abc002")
			assert.Contains(t, output, "abc003")
		})
	})
}
