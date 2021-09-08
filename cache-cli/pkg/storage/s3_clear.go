package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *S3Storage) Clear() error {
	keys, err := s.List()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	output, err := s.client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: &s.bucketName,
		Delete: &types.Delete{
			Objects: cacheKeysToObjectIdentifiers(s.project, keys),
		},
	})

	if err != nil {
		return err
	}

	if len(output.Errors) > 0 {
		firstError := output.Errors[0]
		return fmt.Errorf("clear operation failed, some keys might not have been deleted: %s", *firstError.Message)
	}

	return nil
}

func cacheKeysToObjectIdentifiers(project string, keys []CacheKey) []types.ObjectIdentifier {
	objectIdentifiers := make([]types.ObjectIdentifier, 0)
	for _, key := range keys {
		bucketKey := fmt.Sprintf("%s/%s", project, key.Name)
		identifier := types.ObjectIdentifier{
			Key: &bucketKey,
		}

		objectIdentifiers = append(objectIdentifiers, identifier)
	}

	return objectIdentifiers
}
