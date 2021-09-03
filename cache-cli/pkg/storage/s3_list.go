package storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *S3Storage) List() (*ListResult, error) {
	input := s3.ListObjectsInput{
		Bucket: &s.bucketName,
		Prefix: &s.project,
	}

	output, err := s.client.ListObjects(context.TODO(), &input)
	if err != nil {
		return nil, err
	}

	return createListResult(output.Contents), nil
}

func createListResult(objects []types.Object) *ListResult {
	result := ListResult{
		Keys: make([]CacheKey, 0),
	}

	for _, object := range objects {
		result.Keys = append(result.Keys, CacheKey{
			Name:      *object.Key,
			UpdatedAt: object.LastModified,
		})
	}

	return &result
}
