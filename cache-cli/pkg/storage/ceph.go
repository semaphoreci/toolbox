package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// CephStorage is the cache backend for Semaphore-managed Ceph caches. It is
// S3-compatible storage fronted by a pull-through cache, so it reuses S3Storage
// for everything except:
//   - credentials: short-lived STS credentials obtained via
//     AssumeRoleWithWebIdentity from the cache OIDC token (auto-refreshing);
//   - Store/Restore: aws-cli-shaped requests (single PutObject, sequential
//     8 MiB ranged GETs) plus the x-id middleware, so the pull-through cache
//     can cache GetObject responses.
//
// All other operations (List, HasKey, Delete, Clear, Usage, ...) are inherited
// from S3Storage and run against the same Ceph-configured client.
type CephStorage struct {
	*S3Storage
}

// staticToken is an stscreds.IdentityTokenRetriever that returns the cache OIDC
// token injected by Zebra via SEMAPHORE_CACHE_OIDC_TOKEN.
type staticToken string

func (t staticToken) GetIdentityToken() ([]byte, error) {
	return []byte(t), nil
}

func NewCephStorage(options S3StorageOptions) (*CephStorage, error) {
	token := os.Getenv("SEMAPHORE_CACHE_OIDC_TOKEN")
	roleARN := os.Getenv("SEMAPHORE_CACHE_ROLE_ARN")

	if token == "" {
		return nil, fmt.Errorf("no SEMAPHORE_CACHE_OIDC_TOKEN set")
	}

	if roleARN == "" {
		return nil, fmt.Errorf("no SEMAPHORE_CACHE_ROLE_ARN set")
	}

	if options.URL == "" {
		return nil, fmt.Errorf("no SEMAPHORE_CACHE_S3_URL set")
	}

	region := os.Getenv("SEMAPHORE_CACHE_S3_REGION")
	if region == "" {
		region = "auto"
	}

	cfg, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithRegion(region),
		awsConfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, reg string, opts ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: options.URL, SigningRegion: region}, nil
			}),
		),
	)
	if err != nil {
		return nil, err
	}

	// STS and S3 both target the same Ceph RGW endpoint. The pull-through cache
	// only caches GetObject and passes every other request (including the
	// AssumeRoleWithWebIdentity exchange) straight through to Ceph.
	stsClient := sts.NewFromConfig(cfg)

	credentialsProvider := aws.NewCredentialsCache(
		stscreds.NewWebIdentityRoleProvider(stsClient, roleARN, staticToken(token)),
	)

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.Credentials = credentialsProvider
		o.APIOptions = append(o.APIOptions, removeS3OperationID)
	})

	return &CephStorage{
		S3Storage: &S3Storage{
			Client:        s3Client,
			Bucket:        options.Bucket,
			Project:       options.Project,
			StorageConfig: options.Config,
		},
	}, nil
}

// removeS3OperationID strips the SDK-added `x-id` query parameter (e.g.
// `?x-id=GetObject`) from requests. aws-cli does not send it; removing it keeps
// the request URL stable so the pull-through cache can cache GetObject
// responses.
type removeS3OperationIDMiddleware struct{}

func (m *removeS3OperationIDMiddleware) ID() string {
	return "RemoveS3OperationIDMiddleware"
}

func (m *removeS3OperationIDMiddleware) HandleBuild(
	ctx context.Context,
	in middleware.BuildInput,
	next middleware.BuildHandler,
) (out middleware.BuildOutput, metadata middleware.Metadata, err error) {
	if request, ok := in.Request.(*smithyhttp.Request); ok {
		query := request.URL.Query()
		query.Del("x-id")
		request.URL.RawQuery = query.Encode()
	}

	return next.HandleBuild(ctx, in)
}

func removeS3OperationID(stack *middleware.Stack) error {
	return stack.Build.Add(&removeS3OperationIDMiddleware{}, middleware.After)
}
