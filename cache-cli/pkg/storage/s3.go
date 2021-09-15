package storage

import (
	"context"

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

func NewS3Storage(url, bucket, project string) (*S3Storage, error) {
	if url != "" {
		return createS3StorageUsingEndpoint(bucket, project, url)
	} else {
		return createDefaultS3Storage(bucket, project)
	}
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
