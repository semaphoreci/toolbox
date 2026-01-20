package storage

import (
	"context"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Storage) List() ([]CacheKey, error) {
	prefix := s.Project + "/"
	paginator := s3.NewListObjectsV2Paginator(s.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	})
	keys := make([]CacheKey, 0)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, object := range output.Contents {
			name := strings.TrimPrefix(aws.ToString(object.Key), prefix)
			keys = append(keys, CacheKey{
				Name:           name,
				StoredAt:       object.LastModified,
				LastAccessedAt: object.LastModified,
				Size:           object.Size,
			})
		}
	}
	return s.sortKeys(keys), nil
}

// S3 backend does not support sorting keys by ACCESS_TIME
func (s *S3Storage) sortKeys(keys []CacheKey) []CacheKey {
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
