package cmd

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	assert "github.com/stretchr/testify/assert"
)

func Test__Clear(t *testing.T) {
	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s no keys", backend), func(*testing.T) {
			storage.Clear()

			capturer := utils.CreateOutputCapturer()
			RunClear(clearCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, "Cache is clear.")
		})

		t.Run(fmt.Sprintf("%s with keys", backend), func(*testing.T) {
			storage.Clear()
			tempFile, _ := ioutil.TempFile("/tmp", "*")
			storage.Store("abc001", tempFile.Name())

			capturer := utils.CreateOutputCapturer()
			RunClear(hasKeyCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, "Cache is clear.")
		})
	})
}
