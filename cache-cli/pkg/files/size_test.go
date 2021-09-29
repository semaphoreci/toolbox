package files

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__HumanReadableSize(t *testing.T) {
	t.Run("bytes", func(t *testing.T) {
		assert.Equal(t, "800 B", HumanReadableSize(800))
	})

	t.Run("kilobytes with no floating part", func(t *testing.T) {
		assert.Equal(t, "1.0 kB", HumanReadableSize(1024))
	})

	t.Run("kilobytes with floating part", func(t *testing.T) {
		assert.Equal(t, "38.2 kB", HumanReadableSize(1024*38+256))
	})

	t.Run("megabytes with no floating part", func(t *testing.T) {
		assert.Equal(t, "5.0 MB", HumanReadableSize(1024*1024*5))
	})

	t.Run("megabytes with floating part", func(t *testing.T) {
		assert.Equal(t, "5.2 MB", HumanReadableSize(1024*1024*5+1024*256))
	})

	t.Run("gigabytes with no floating part", func(t *testing.T) {
		assert.Equal(t, "5.0 GB", HumanReadableSize(1024*1024*1024*5))
	})

	t.Run("gigabytes with floating part", func(t *testing.T) {
		assert.Equal(t, "5.2 GB", HumanReadableSize(1024*1024*1024*5+1024*1024*256))
	})
}
