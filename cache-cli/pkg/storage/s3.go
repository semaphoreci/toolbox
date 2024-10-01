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

func NewS3Storage(options S3StorageOptions) (*S3Storage, error) {
	if options.URL != "" {
		return createS3StorageUsingEndpoint(options.Bucket, options.Project, options.URL, options.Config)
	}

	return createDefaultS3Storage(options.Bucket, options.Project, options.Config)
}

func createDefaultS3Storage(s3Bucket, project string, storageConfig StorageConfig) (*S3Storage, error) {
	var config aws.Config
	var err error

	// Using EC2 metadata service to retrieve credentials for the instance profile
	if os.Getenv("SEMAPHORE_CACHE_USE_EC2_INSTANCE_PROFILE") == "true" {
		log.Infof("Using EC2 instance profile.")
		config, err = awsConfig.LoadDefaultConfig(
			context.TODO(),
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
		config, err = awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithSharedConfigProfile(profile))
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
	config, err = awsConfig.LoadDefaultConfig(context.TODO())
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
	options := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRegion("auto"),
		awsConfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: s3Url,
				}, nil
			}),
		),
	}

	// If a key/secret pair is passed, we use them.
	// Otherwise, we just rely on the default configuration methods
	// used by LoadDefaultConfig(), for example,
	// AWS_SECRET_ACCESS_KEY and AWS_ACCESS_KEY_ID environment variables.
	s3Key := os.Getenv("SEMAPHORE_CACHE_S3_KEY")
	s3Secret := os.Getenv("SEMAPHORE_CACHE_S3_SECRET")
	if s3Key != "" && s3Secret != "" {
		options = append(options, awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s3Key, s3Secret, ""),
		))
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), options...)
	if err != nil {
		return nil, err
	}

	return &S3Storage{
		Bucket:        s3Bucket,
		Project:       project,
		StorageConfig: storageConfig,
		Client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true
		}),
	}, nil
}

func (s *S3Storage) Config() StorageConfig {
	return s.StorageConfig
}
