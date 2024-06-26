package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)

	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.Bucket,
		Key:    &bucketKey,
	})

	return err
}
