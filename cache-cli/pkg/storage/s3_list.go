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
	output, err := s.client.ListObjects(context.TODO(), s.listObjectsInput(nil))
	if err != nil {
		return nil, err
	}

	keys := make([]CacheKey, 0)
	keys = s.appendToListResult(keys, output.Contents)

	for output.IsTruncated {
		nextMarker := findNextMarker(output)
		output, err = s.client.ListObjects(context.TODO(), s.listObjectsInput(&nextMarker))
		if err != nil {
			return nil, err
		}

		keys = s.appendToListResult(keys, output.Contents)
	}

	return keys, nil
}

func (s *S3Storage) listObjectsInput(nextMarker *string) *s3.ListObjectsInput {
	if nextMarker != nil {
		return &s3.ListObjectsInput{
			Bucket: &s.bucketName,
			Prefix: &s.project,
			Marker: nextMarker,
		}
	} else {
		return &s3.ListObjectsInput{
			Bucket: &s.bucketName,
			Prefix: &s.project,
		}
	}
}

func (s *S3Storage) appendToListResult(keys []CacheKey, objects []types.Object) []CacheKey {
	for _, object := range objects {
		keyWithoutProject := strings.ReplaceAll(*object.Key, fmt.Sprintf("%s/", s.project), "")
		keys = append(keys, CacheKey{
			Name:     keyWithoutProject,
			StoredAt: object.LastModified,
			Size:     object.Size,
		})
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].StoredAt.After(*keys[j].StoredAt)
	})

	return keys
}

func findNextMarker(output *s3.ListObjectsOutput) string {
	if output.NextMarker != nil {
		return *output.NextMarker
	} else {
		contents := output.Contents
		lastElement := contents[len(contents)-1]
		return *lastElement.Key
	}
}
