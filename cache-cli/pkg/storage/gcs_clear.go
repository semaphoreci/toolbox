package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func (s *GCSStorage) Clear() error {
	it := s.Bucket.Objects(context.TODO(), &storage.Query{Prefix: ""})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		err = s.Bucket.Object(attrs.Name).Delete(context.TODO())
		if err != nil {
			return err
		}
	}

	return nil
}
