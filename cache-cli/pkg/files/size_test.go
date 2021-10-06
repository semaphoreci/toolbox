package files

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__HumanReadableSize(t *testing.T) {
	t.Run("bytes", func(t *testing.T) {
		assert.Equal(t, "800.0", HumanReadableSize(800))
	})

	t.Run("kilobytes", func(t *testing.T) {
		assert.Equal(t, "100.0K", HumanReadableSize(1024*100))
	})

	t.Run("kilobytes with remainder", func(t *testing.T) {
		assert.Equal(t, "100.1K", HumanReadableSize(1024*100+128))
	})

	t.Run("megabytes", func(t *testing.T) {
		assert.Equal(t, "5.0M", HumanReadableSize(1024*1024*5))
	})

	t.Run("megabytes with remainder", func(t *testing.T) {
		assert.Equal(t, "5.1M", HumanReadableSize(1024*1024*5+1024*128))
	})

	t.Run("gigabytes", func(t *testing.T) {
		assert.Equal(t, "5.0G", HumanReadableSize(1024*1024*1024*5))
	})

	t.Run("gigabytes with remainder", func(t *testing.T) {
		assert.Equal(t, "5.1G", HumanReadableSize(1024*1024*1024*5+1024*1024*128))
	})
}
