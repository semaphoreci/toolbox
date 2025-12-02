package compression

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
)

// GzipDecompress takes a reader and returns a reader that handles both compressed and uncompressed data
func GzipDecompress(reader io.Reader) (io.Reader, func() error, error) {
	isCompressed, reader, err := IsGzipCompressed(reader)
	if err != nil {
		return nil, nil, err
	}

	if !isCompressed {
		// Return original reader with no-op close function
		return reader, func() error { return nil }, nil
	}

	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		logger.Error("Creating gzip reader failed: %v", err)
		return nil, nil, err
	}

	return gzReader, gzReader.Close, nil
}

// IsGzipCompressed checks if data is gzip compressed and returns a reader with the header bytes intact
func IsGzipCompressed(reader io.Reader) (bool, io.Reader, error) {
	header := make([]byte, 2)
	n, err := io.ReadFull(reader, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, reader, err
	}

	if n < 2 {
		// Not enough data to determine, treat as uncompressed
		reader = io.MultiReader(bytes.NewReader(header[:n]), reader)
		return false, reader, nil
	}

	// Check if the header matches GZIP magic numbers
	isCompressed := header[0] == 0x1f && header[1] == 0x8b

	// Combine the read bytes with the original reader
	reader = io.MultiReader(bytes.NewReader(header), reader)

	return isCompressed, reader, nil
}

// GzipCompress compresses data using gzip
func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return data, err
	}

	err = writer.Close()
	if err != nil {
		return data, err
	}

	return buf.Bytes(), nil
}

// IsGzipCompressedBytes checks if a byte slice is gzip compressed
func IsGzipCompressedBytes(bytes []byte) bool {
	if len(bytes) < 2 {
		return false
	}
	return bytes[0] == 0x1f && bytes[1] == 0x8b
}
