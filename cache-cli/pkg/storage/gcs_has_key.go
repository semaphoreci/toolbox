package storage

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
)

func (s *GCSStorage) HasKey(key string) (bool, error) {
	gcsKey := fmt.Sprintf("%s/%s", s.Project, key)
	_, err := s.Bucket.Object(gcsKey).Attrs(context.TODO())
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
