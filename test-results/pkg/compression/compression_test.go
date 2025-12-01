package compression_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/semaphoreci/toolbox/test-results/pkg/compression"
	"github.com/stretchr/testify/assert"
)

func TestIsGzipCompressedBytes(t *testing.T) {
	testCases := []struct {
		Name  string
		Input []byte
		Want  bool
	}{
		{
			Name:  "Empty input",
			Input: []byte{},
			Want:  false,
		},
		{
			Name:  "Single byte",
			Input: []byte{0x1f},
			Want:  false,
		},
		{
			Name:  "Gzip header",
			Input: []byte{0x1f, 0x8b},
			Want:  true,
		},
		{
			Name:  "Gzip compressed empty data",
			Input: []byte{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff},
			Want:  true,
		},
		{
			Name:  "Not gzip",
			Input: []byte{0x50, 0x4b}, // ZIP header
			Want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Want, compression.IsGzipCompressedBytes(tc.Input))
		})
	}
}

func TestIsGzipCompressed(t *testing.T) {
	testCases := []struct {
		Name  string
		Input []byte
		Want  bool
	}{
		{
			Name:  "Empty input",
			Input: []byte{},
			Want:  false,
		},
		{
			Name:  "Gzip header",
			Input: []byte{0x1f, 0x8b, 0x8, 0x0},
			Want:  true,
		},
		{
			Name:  "Not gzip",
			Input: []byte("plain text"),
			Want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			isCompressed, reader, err := compression.IsGzipCompressed(bytes.NewReader(tc.Input))
			assert.NoError(t, err)
			assert.Equal(t, tc.Want, isCompressed)

			// Verify we can still read the original data
			data, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.Equal(t, tc.Input, data)
		})
	}
}

func TestGzipCompressDecompress(t *testing.T) {
	testCases := []struct {
		Name  string
		Input string
	}{
		{
			Name:  "Empty input",
			Input: "",
		},
		{
			Name:  "Simple text",
			Input: "Hello, World!",
		},
		{
			Name:  "Large text",
			Input: string(make([]byte, 10000)),
		},
		{
			Name:  "Binary data",
			Input: string([]byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0xff}),
		},
		{
			Name:  "Unicode text",
			Input: "Hello 世界 🌍",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Test GzipCompress
			compressed, err := compression.GzipCompress([]byte(tc.Input))
			assert.NoError(t, err)
			assert.True(t, compression.IsGzipCompressedBytes(compressed))

			// Test GzipDecompress with reader
			reader, closeFunc, err := compression.GzipDecompress(bytes.NewReader(compressed))
			assert.NoError(t, err)
			defer closeFunc()

			decompressed, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.Equal(t, tc.Input, string(decompressed))
		})
	}
}

func TestGzipDecompressWithUncompressedData(t *testing.T) {
	input := []byte("This is uncompressed data")

	reader, closeFunc, err := compression.GzipDecompress(bytes.NewReader(input))
	assert.NoError(t, err)
	defer closeFunc()

	// Should return the original data since it's not compressed
	output, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, input, output)
}

func TestGzipDecompressWithCompressedData(t *testing.T) {
	original := []byte("This will be compressed")

	// First compress it
	compressed, err := compression.GzipCompress(original)
	assert.NoError(t, err)

	// Now decompress it
	reader, closeFunc, err := compression.GzipDecompress(bytes.NewReader(compressed))
	assert.NoError(t, err)
	defer closeFunc()

	decompressed, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, original, decompressed)
}
