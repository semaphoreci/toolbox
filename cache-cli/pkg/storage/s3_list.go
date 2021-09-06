package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *S3Storage) List() ([]CacheKey, error) {
	input := s3.ListObjectsInput{
		Bucket: &s.bucketName,
		Prefix: &s.project,
	}

	output, err := s.client.ListObjects(context.TODO(), &input)
	if err != nil {
		return nil, err
	}

	return createListResult(s.project, output.Contents), nil
}

func createListResult(project string, objects []types.Object) []CacheKey {
	keys := make([]CacheKey, 0)

	for _, object := range objects {
		keyWithoutProject := strings.ReplaceAll(*object.Key, fmt.Sprintf("%s/", project), "")
		keys = append(keys, CacheKey{
			Name:     keyWithoutProject,
			StoredAt: object.LastModified,
		})
	}

	return keys
}
