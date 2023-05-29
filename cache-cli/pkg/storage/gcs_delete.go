package storage

import (
	"context"
	"errors"
	"fmt"

	gcs "cloud.google.com/go/storage"
)

func (s *GCSStorage) Delete(key string) error {
	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)
	err := s.Bucket.Object(bucketKey).Delete(context.TODO())
	if errors.Is(err, gcs.ErrObjectNotExist) {
		return nil
	}

	return err
}
