package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Storage) Restore(key string, writer io.Writer) (int64, error) {
	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)
	response, err := s.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &bucketKey,
	})

	if err != nil {
		return 0, err
	}

	return io.Copy(writer, response.Body)
}
