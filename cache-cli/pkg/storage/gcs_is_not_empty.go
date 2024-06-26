package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func (s *GCSStorage) IsNotEmpty(ctx context.Context) (bool, error) {
	it := s.Bucket.Objects(ctx, &storage.Query{Prefix: s.Project})

	_, err := it.Next()
	if err == iterator.Done {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
