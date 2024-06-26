package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	assert "github.com/stretchr/testify/assert"
)

func Test__IsNotEmpty(t *testing.T) {
	ctx := context.TODO()
	runTestForAllBackends(ctx, t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s cache is empty", backend), func(*testing.T) {
			storage.Clear(ctx)
			assert.False(t, RunIsNotEmpty(isNotEmptyCmd, []string{}))
		})

		t.Run(fmt.Sprintf("%s cache is not empty", backend), func(*testing.T) {
			storage.Clear(ctx)
			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			storage.Store(ctx, "abc001", tempFile.Name())

			assert.True(t, RunIsNotEmpty(isNotEmptyCmd, []string{}))
		})
	})
}
