package storage

import (
	"context"
)

func (s *GCSStorage) Delete(key string) error {
	return s.Bucket.Object(key).Delete(context.TODO())
}
