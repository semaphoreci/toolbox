package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func (s *GCSStorage) IsNotEmpty() (bool, error) {
	it := s.Bucket.Objects(context.TODO(), &storage.Query{Prefix: ""})

	_, err := it.Next()
	if err == iterator.Done {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
