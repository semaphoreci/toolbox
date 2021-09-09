package cmd

import (
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	assert "github.com/stretchr/testify/assert"
)

func Test__Usage(t *testing.T) {
	storage, err := storage.InitStorage()
	assert.Nil(t, err)

	storage.Clear()

	capturer := utils.CreateOutputCapturer()
	RunUsage(usageCmd, []string{})
	output := capturer.Done()

	assert.Contains(t, output, "FREE SPACE: (unlimited)")
	assert.Contains(t, output, "USED SPACE: 0 B")
}
