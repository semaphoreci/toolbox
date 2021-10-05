package cmd

import (
	"io/ioutil"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	assert "github.com/stretchr/testify/assert"
)

func Test__Clear(t *testing.T) {
	storage, err := storage.InitStorage()
	assert.Nil(t, err)

	t.Run("no keys", func(*testing.T) {
		storage.Clear()

		capturer := utils.CreateOutputCapturer()
		RunClear(clearCmd, []string{})
		output := capturer.Done()

		assert.Contains(t, output, "Cache is clear.")
	})

	t.Run("with keys", func(*testing.T) {
		storage.Clear()
		tempFile, _ := ioutil.TempFile("/tmp", "*")
		storage.Store("abc001", tempFile.Name())

		capturer := utils.CreateOutputCapturer()
		RunClear(hasKeyCmd, []string{})
		output := capturer.Done()

		assert.Contains(t, output, "Cache is clear.")
	})
}
