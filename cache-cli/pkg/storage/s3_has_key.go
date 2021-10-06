package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

func (s *S3Storage) HasKey(key string) (bool, error) {
	s3Key := fmt.Sprintf("%s/%s", s.Project, key)
	input := s3.HeadObjectInput{
		Bucket: &s.Bucket,
		Key:    &s3Key,
	}

	_, err := s.Client.HeadObject(context.TODO(), &input)
	if err != nil {
		var apiErr *smithy.GenericAPIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
