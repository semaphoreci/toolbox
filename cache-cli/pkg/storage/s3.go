package storage

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
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

func NewS3Storage(ctx context.Context, options S3StorageOptions) (*S3Storage, error) {
	if options.URL != "" {
		return createS3StorageUsingEndpoint(ctx, options.Bucket, options.Project, options.URL, options.Config)
	}

	return createDefaultS3Storage(ctx, options.Bucket, options.Project, options.Config)
}

func createDefaultS3Storage(ctx context.Context, s3Bucket, project string, storageConfig StorageConfig) (*S3Storage, error) {
	var config aws.Config
	var err error

	// Using EC2 metadata service to retrieve credentials for the instance profile
	if os.Getenv("SEMAPHORE_CACHE_USE_EC2_INSTANCE_PROFILE") == "true" {
		log.Infof("Using EC2 instance profile.")
		config, err = awsConfig.LoadDefaultConfig(
			ctx,
			awsConfig.WithCredentialsProvider(ec2rolecreds.New()),
			awsConfig.WithEC2IMDSRegion(),
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

	// Using an existing profile configured in one of the default configuration files.
	profile := os.Getenv("SEMAPHORE_CACHE_AWS_PROFILE")
	if profile != "" {
		log.Infof("Using '%s' AWS profile.", profile)
		config, err = awsConfig.LoadDefaultConfig(ctx, awsConfig.WithSharedConfigProfile(profile))
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

	// No special configuration, just follow the default chain
	config, err = awsConfig.LoadDefaultConfig(ctx)
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

func createS3StorageUsingEndpoint(ctx context.Context, s3Bucket, project, s3Url string, storageConfig StorageConfig) (*S3Storage, error) {
	resolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: s3Url,
		}, nil
	})

	creds := credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", "")
	cfg, err := awsConfig.LoadDefaultConfig(ctx,
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
