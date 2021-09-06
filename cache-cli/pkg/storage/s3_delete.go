package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Storage) Delete(key string) error {
	bucketKey := fmt.Sprintf("%s/%s", s.project, key)

	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &s.bucketName,
		Key:    &bucketKey,
	})

	return err
}
