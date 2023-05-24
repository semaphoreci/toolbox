package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func (s *GCSStorage) Restore(key string) (*os.File, error) {
	tempFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-*", key))
	if err != nil {
		return nil, err
	}

	reader, err := s.Bucket.Object(key).NewReader(context.TODO())
	if err != nil {
		_ = tempFile.Close()
		return nil, err
	}

	defer reader.Close()

	_, err = io.Copy(tempFile, reader)
	if err != nil {
		_ = tempFile.Close()
		return nil, err
	}

	return tempFile, tempFile.Close()
}
