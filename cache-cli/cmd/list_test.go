package cmd

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	assert "github.com/stretchr/testify/assert"
)

func Test__List(t *testing.T) {
	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s no keys", backend), func(*testing.T) {
			storage.Clear()

			capturer := utils.CreateOutputCapturer()
			RunList(listCmd, []string{""})
			output := capturer.Done()

			assert.Contains(t, output, "Cache is empty.")
		})

		t.Run(fmt.Sprintf("%s with keys", backend), func(*testing.T) {
			storage.Clear()
			tempFile, _ := ioutil.TempFile("/tmp", "*")
			storage.Store("abc001", tempFile.Name())
			storage.Store("abc002", tempFile.Name())
			storage.Store("abc003", tempFile.Name())

			capturer := utils.CreateOutputCapturer()
			RunList(listCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, "abc001")
			assert.Contains(t, output, "abc002")
			assert.Contains(t, output, "abc003")
		})
	})
}
