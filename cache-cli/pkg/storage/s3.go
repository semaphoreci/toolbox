package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	project    string
}

func NewS3Storage() (*S3Storage, error) {
	project := os.Getenv("SEMAPHORE_PROJECT_NAME")
	if project == "" {
		return nil, fmt.Errorf("no SEMAPHORE_PROJECT_NAME set")
	}

	s3Bucket := os.Getenv("SEMAPHORE_CACHE_S3_BUCKET")
	if s3Bucket == "" {
		return nil, fmt.Errorf("no SEMAPHORE_CACHE_S3_BUCKET set")
	}

	s3Url := os.Getenv("SEMAPHORE_CACHE_S3_URL")
	if s3Url != "" {
		return createS3StorageUsingEndpoint(s3Bucket, project, s3Url)
	}

	return createDefaultS3Storage(s3Bucket, project)
}

func createDefaultS3Storage(s3Bucket, project string) (*S3Storage, error) {
	config, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return &S3Storage{
		client:     s3.NewFromConfig(config),
		bucketName: s3Bucket,
		project:    project,
	}, nil
}

func createS3StorageUsingEndpoint(s3Bucket, project, s3Url string) (*S3Storage, error) {
	resolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: s3Url,
		}, nil
	})

	creds := credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", "")
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithCredentialsProvider(creds),
		awsConfig.WithEndpointResolver(resolver),
	)

	if err != nil {
		return nil, err
	}

	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Storage{
		client:     svc,
		bucketName: s3Bucket,
		project:    project,
	}, nil
}
