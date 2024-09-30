package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *S3Storage) List() ([]CacheKey, error) {
	output, err := s.Client.ListObjects(context.TODO(), s.listObjectsInput(nil))
	if err != nil {
		return nil, err
	}

	keys := make([]CacheKey, 0)
	keys = s.appendToListResult(keys, output.Contents)

	for output.IsTruncated != nil && *output.IsTruncated {
		nextMarker := findNextMarker(output)
		output, err = s.Client.ListObjects(context.TODO(), s.listObjectsInput(&nextMarker))
		if err != nil {
			return nil, err
		}

		keys = s.appendToListResult(keys, output.Contents)
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

func (s *S3Storage) listObjectsInput(nextMarker *string) *s3.ListObjectsInput {
	if nextMarker != nil {
		return &s3.ListObjectsInput{
			Bucket: &s.Bucket,
			Prefix: &s.Project,
			Marker: nextMarker,
		}
	}

	return &s3.ListObjectsInput{
		Bucket: &s.Bucket,
		Prefix: &s.Project,
	}
}

func (s *S3Storage) appendToListResult(keys []CacheKey, objects []types.Object) []CacheKey {
	for _, object := range objects {
		keyWithoutProject := strings.ReplaceAll(*object.Key, fmt.Sprintf("%s/", s.Project), "")
		keys = append(keys, CacheKey{
			Name:           keyWithoutProject,
			StoredAt:       object.LastModified,
			LastAccessedAt: object.LastModified,
			Size:           *object.Size,
		})
	}

	return keys
}

func findNextMarker(output *s3.ListObjectsOutput) string {
	if output.NextMarker != nil {
		return *output.NextMarker
	}

	contents := output.Contents
	lastElement := contents[len(contents)-1]
	return *lastElement.Key
}
