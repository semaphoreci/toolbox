package storage

import (
	"context"
	"fmt"
)

func (s *GCSStorage) Delete(key string) error {
	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)
	return s.Bucket.Object(bucketKey).Delete(context.TODO())
}
