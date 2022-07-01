package storage

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	Client        *s3.Client
	Bucket        string
	Project       string
	StorageConfig StorageConfig
}

type S3StorageOptions struct {
	URL     string
	Bucket  string
	Project string
	Config  StorageConfig
}

func NewS3Storage(options S3StorageOptions) (*S3Storage, error) {
	if options.URL != "" {
		return createS3StorageUsingEndpoint(options.Bucket, options.Project, options.URL, options.Config)
	}

	return createDefaultS3Storage(options.Bucket, options.Project, options.Config)
}

func createDefaultS3Storage(s3Bucket, project string, storageConfig StorageConfig) (*S3Storage, error) {
	config, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithSharedConfigProfile(
			os.Getenv("SEMAPHORE_CACHE_AWS_PROFILE"),
		),
	)

	if err != nil {
		return nil, err
	}

	return &S3Storage{
		Client:        s3.NewFromConfig(config),
		Bucket:        s3Bucket,
		Project:       project,
		StorageConfig: storageConfig,
	}, nil
}

func createS3StorageUsingEndpoint(s3Bucket, project, s3Url string, storageConfig StorageConfig) (*S3Storage, error) {
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
		Client:        svc,
		Bucket:        s3Bucket,
		Project:       project,
		StorageConfig: storageConfig,
	}, nil
}

func (s *S3Storage) Config() StorageConfig {
	return s.StorageConfig
}
