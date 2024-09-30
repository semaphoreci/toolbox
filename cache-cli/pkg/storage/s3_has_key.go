package storage

import (
	"context"
	"errors"
	"fmt"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Storage) HasKey(key string) (bool, error) {
	s3Key := fmt.Sprintf("%s/%s", s.Project, key)
	input := s3.HeadObjectInput{
		Bucket: &s.Bucket,
		Key:    &s3Key,
	}

	_, err := s.Client.HeadObject(context.TODO(), &input)
	if err != nil {
		var apiErr *awshttp.ResponseError
		if errors.As(err, &apiErr) && apiErr.HTTPStatusCode() == 404 {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
