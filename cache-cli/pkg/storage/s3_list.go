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
	input := s3.ListObjectsInput{
		Bucket: &s.Bucket,
		Prefix: &s.Project,
	}

	output, err := s.Client.ListObjects(context.TODO(), &input)
	if err != nil {
		return nil, err
	}

	return createListResult(s.Project, output.Contents), nil
}

func createListResult(project string, objects []types.Object) []CacheKey {
	keys := make([]CacheKey, 0)

	for _, object := range objects {
		keyWithoutProject := strings.ReplaceAll(*object.Key, fmt.Sprintf("%s/", project), "")
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
