package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func (s *GCSStorage) List() ([]CacheKey, error) {
	it := s.Bucket.Objects(context.TODO(), &storage.Query{Prefix: s.Project})

	keys := make([]CacheKey, 0)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		keys = s.appendToListResult(keys, attrs)
	}

	return s.sortKeys(keys), nil
}

// S3 backend does not support sorting keys by ACCESS_TIME
func (s *GCSStorage) sortKeys(keys []CacheKey) []CacheKey {
	switch s.Config().SortKeysBy {
	case SortBySize:
		sort.SliceStable(keys, func(i, j int) bool {
			return keys[i].Size > keys[j].Size
		})
	default:
		sort.SliceStable(keys, func(i, j int) bool {
			return keys[i].StoredAt.After(*keys[j].StoredAt)
		})
	}

	return keys
}

func (s *GCSStorage) appendToListResult(keys []CacheKey, object *storage.ObjectAttrs) []CacheKey {
	keyWithoutProject := strings.ReplaceAll(object.Name, fmt.Sprintf("%s/", s.Project), "")
	keys = append(keys, CacheKey{
		Name:           keyWithoutProject,
		StoredAt:       &object.Updated,
		LastAccessedAt: &object.Updated,
		Size:           object.Size,
	})

	return keys
}
