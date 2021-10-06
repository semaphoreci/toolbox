package files

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__HumanReadableSize(t *testing.T) {
	t.Run("bytes", func(t *testing.T) {
		assert.Equal(t, "800B", HumanReadableSize(800))
	})

	t.Run("kilobytes", func(t *testing.T) {
		assert.Equal(t, "100K", HumanReadableSize(1024*100))
	})

	t.Run("megabytes with no floating part", func(t *testing.T) {
		assert.Equal(t, "5M", HumanReadableSize(1024*1024*5))
	})

	t.Run("gigabytes with no floating part", func(t *testing.T) {
		assert.Equal(t, "5G", HumanReadableSize(1024*1024*1024*5))
	})
}
