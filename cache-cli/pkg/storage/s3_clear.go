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

	// the s3 DeleteObjects operation only allows up to 1000 keys to be used
	chunks := createChunks(keys, 1000)

	for _, chunk := range chunks {
		err := s.deleteChunk(chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *S3Storage) deleteChunk(keys []CacheKey) error {
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

func createChunks(keys []CacheKey, chunkSize int) [][]CacheKey {
	var chunks [][]CacheKey
	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize

		if end > len(keys) {
			end = len(keys)
		}

		chunks = append(chunks, keys[i:end])
	}

	return chunks
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
