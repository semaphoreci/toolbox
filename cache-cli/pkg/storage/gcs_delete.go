package storage

import (
	"context"
	"errors"
	"fmt"

	gcs "cloud.google.com/go/storage"
)

func (s *GCSStorage) Delete(ctx context.Context, key string) error {
	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)
	err := s.Bucket.Object(bucketKey).Delete(ctx)
	if errors.Is(err, gcs.ErrObjectNotExist) {
		return nil
	}

	return err
}
