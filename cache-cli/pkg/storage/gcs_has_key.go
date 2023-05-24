package storage

import (
	"context"
	"errors"

	"cloud.google.com/go/storage"
)

func (s *GCSStorage) HasKey(key string) (bool, error) {
	_, err := s.Bucket.Object(key).Attrs(context.TODO())
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
