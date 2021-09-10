package cmd

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	assert "github.com/stretchr/testify/assert"
)

func Test__Store(t *testing.T) {
	storage, err := storage.InitStorage()
	assert.Nil(t, err)

	t.Run("wrong number of arguments", func(t *testing.T) {
		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{"key", "value", "extra-bad-argument"})
		output := capturer.Done()

		assert.Contains(t, output, "Wrong number of arguments")
	})

	t.Run("using key and invalid path", func(*testing.T) {
		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{"abc001", "/tmp/this-path-does-not-exist"})
		output := capturer.Done()

		assert.Contains(t, output, "Path /tmp/this-path-does-not-exist does not exist")
	})

	t.Run("using key and valid path", func(*testing.T) {
		storage.Clear()
		tempDir, _ := ioutil.TempDir("/tmp", "*")
		ioutil.TempFile(tempDir, "*")

		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{"abc002", tempDir})
		output := capturer.Done()

		assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc002'", tempDir))
		assert.Contains(t, output, "Upload complete")
	})

	t.Run("using duplicate key", func(*testing.T) {
		storage.Clear()
		tempDir, _ := ioutil.TempDir("/tmp", "*")
		ioutil.TempFile(tempDir, "*")

		// Storing key for the first time
		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{"abc003", tempDir})
		output := capturer.Done()
		assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc003'", tempDir))
		assert.Contains(t, output, "Upload complete")

		// Storing key for the second time
		capturer = utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{"abc003", tempDir})
		output = capturer.Done()
		assert.Contains(t, output, "Key 'abc003' already exists")
	})
}

func Test_AutomaticStore(t *testing.T) {
	// TODO
}
