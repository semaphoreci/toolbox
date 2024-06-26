package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func (s *GCSStorage) Clear(ctx context.Context) error {
	it := s.Bucket.Objects(ctx, &storage.Query{Prefix: s.Project})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		err = s.Bucket.Object(attrs.Name).Delete(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
