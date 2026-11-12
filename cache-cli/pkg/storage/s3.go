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
	config, err := loadDefaultS3Config()
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

func loadDefaultS3Config() (aws.Config, error) {
	// Using EC2 metadata service to retrieve credentials for the instance profile
	if os.Getenv("SEMAPHORE_CACHE_USE_EC2_INSTANCE_PROFILE") == "true" {
		log.Infof("Using EC2 instance profile.")
		return awsConfig.LoadDefaultConfig(
			context.TODO(),
			awsConfig.WithCredentialsProvider(ec2rolecreds.New()),
			awsConfig.WithEC2IMDSRegion(),
		)
	}

	// Using an existing profile configured in one of the default configuration files.
	profile := os.Getenv("SEMAPHORE_CACHE_AWS_PROFILE")
	if profile != "" {
		log.Infof("Using '%s' AWS profile.", profile)
		return awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithSharedConfigProfile(profile))
	}

	// Using a key and secret from the cache variables, with an optional session token.
	if provider, ok := staticCredentialsFromEnv(); ok {
		log.Infof("Using S3 credentials from SEMAPHORE_CACHE_S3_KEY and SEMAPHORE_CACHE_S3_SECRET.")
		options := []func(*awsConfig.LoadOptions) error{awsConfig.WithCredentialsProvider(provider)}
		if s3Region := os.Getenv("SEMAPHORE_CACHE_S3_REGION"); s3Region != "" {
			options = append(options, awsConfig.WithRegion(s3Region))
		}

		return awsConfig.LoadDefaultConfig(context.TODO(), options...)
	}

	// No special configuration, just follow the default chain
	return awsConfig.LoadDefaultConfig(context.TODO())
}

func createS3StorageUsingEndpoint(s3Bucket, project, s3Url string, storageConfig StorageConfig) (*S3Storage, error) {
	cfg, err := loadS3ConfigUsingEndpoint(s3Url)
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

func loadS3ConfigUsingEndpoint(s3Url string) (aws.Config, error) {
	s3Region := os.Getenv("SEMAPHORE_CACHE_S3_REGION")
	if s3Region == "" {
		s3Region = "auto"
	}

	options := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRegion(s3Region),
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
	if provider, ok := staticCredentialsFromEnv(); ok {
		options = append(options, awsConfig.WithCredentialsProvider(provider))
	}

	return awsConfig.LoadDefaultConfig(context.TODO(), options...)
}

// staticCredentialsFromEnv reads the key, secret and optional session token from
// the cache variables only, never from the generic AWS ones, which may hold
// credentials for the job's own tools.
func staticCredentialsFromEnv() (credentials.StaticCredentialsProvider, bool) {
	key := os.Getenv("SEMAPHORE_CACHE_S3_KEY")
	secret := os.Getenv("SEMAPHORE_CACHE_S3_SECRET")
	if key == "" || secret == "" {
		return credentials.StaticCredentialsProvider{}, false
	}

	return credentials.NewStaticCredentialsProvider(key, secret, os.Getenv("SEMAPHORE_CACHE_S3_SESSION_TOKEN")), true
}

func (s *S3Storage) Config() StorageConfig {
	return s.StorageConfig
}
