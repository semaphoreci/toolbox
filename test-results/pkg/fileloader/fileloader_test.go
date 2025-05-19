package fileloader

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {

	reader := bytes.NewReader([]byte(`Some data`))

	_, found1 := Load("foo", reader)
	_, found2 := Load("foo", reader)
	_, found3 := Load("bar", reader)

	assert.Equal(t, false, found1, "Decoders should be the same")
	assert.Equal(t, true, found2, "Decoders should be the same")
	assert.Equal(t, false, found3, "Decoders should be the same")
}
