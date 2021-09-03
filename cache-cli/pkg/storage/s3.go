package storage

import (
	"context"
	"fmt"
	"os"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	project    string
}

func NewS3Storage() (*S3Storage, error) {
	config, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	s3Bucket := os.Getenv("SEMAPHORE_CACHE_S3_BUCKET")
	if s3Bucket == "" {
		return nil, fmt.Errorf("no SEMAPHORE_CACHE_S3_BUCKET set")
	}

	project := os.Getenv("SEMAPHORE_PROJECT_NAME")
	if project == "" {
		return nil, fmt.Errorf("no SEMAPHORE_PROJECT_NAME set")
	}

	s3Storage := S3Storage{
		client:     s3.NewFromConfig(config),
		bucketName: s3Bucket,
		project:    project,
	}

	return &s3Storage, nil
}
