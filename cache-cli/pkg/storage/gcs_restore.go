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

	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)
	reader, err := s.Bucket.Object(bucketKey).NewReader(context.TODO())
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
