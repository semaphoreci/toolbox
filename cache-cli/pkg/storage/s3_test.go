package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	assert "github.com/stretchr/testify/assert"
)

const testS3URL = "http://s3.example.com:9000"

func Test__StaticCredentialsFromEnv(t *testing.T) {
	t.Run("key, secret and session token", func(t *testing.T) {
		isolateS3Env(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "cache-token")

		provider, ok := staticCredentialsFromEnv()
		if assert.True(t, ok) {
			assert.Equal(t, "cache-key", provider.Value.AccessKeyID)
			assert.Equal(t, "cache-secret", provider.Value.SecretAccessKey)
			assert.Equal(t, "cache-token", provider.Value.SessionToken)
		}
	})

	t.Run("session token is optional and not read from AWS_SESSION_TOKEN", func(t *testing.T) {
		isolateS3Env(t)
		setGenericAWSEnv(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "")

		provider, ok := staticCredentialsFromEnv()
		if assert.True(t, ok) {
			assert.Equal(t, "cache-key", provider.Value.AccessKeyID)
			assert.Equal(t, "cache-secret", provider.Value.SecretAccessKey)
			assert.Empty(t, provider.Value.SessionToken)
		}
	})

	t.Run("key or secret missing", func(t *testing.T) {
		isolateS3Env(t)
		setGenericAWSEnv(t)

		setCacheS3Credentials(t, "cache-key", "", "cache-token")
		_, ok := staticCredentialsFromEnv()
		assert.False(t, ok)

		setCacheS3Credentials(t, "", "cache-secret", "cache-token")
		_, ok = staticCredentialsFromEnv()
		assert.False(t, ok)
	})
}

func Test__LoadDefaultS3Config(t *testing.T) {
	t.Run("cache credentials and region are used over the generic AWS variables", func(t *testing.T) {
		isolateS3Env(t)
		setGenericAWSEnv(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "cache-token")
		t.Setenv("SEMAPHORE_CACHE_S3_REGION", "eu-west-1")

		cfg, err := loadDefaultS3Config()
		if assert.Nil(t, err) {
			assert.Equal(t, "eu-west-1", cfg.Region)
			assert.Nil(t, cfg.EndpointResolverWithOptions)
			assertS3Credentials(t, cfg, "cache-key", "cache-secret", "cache-token")
		}
	})

	t.Run("cache credentials without a session token", func(t *testing.T) {
		isolateS3Env(t)
		setGenericAWSEnv(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "")
		t.Setenv("SEMAPHORE_CACHE_S3_REGION", "eu-west-1")

		cfg, err := loadDefaultS3Config()
		if assert.Nil(t, err) {
			assert.Equal(t, "eu-west-1", cfg.Region)
			assertS3Credentials(t, cfg, "cache-key", "cache-secret", "")
		}
	})

	t.Run("region from the default chain when SEMAPHORE_CACHE_S3_REGION is not set", func(t *testing.T) {
		isolateS3Env(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "cache-token")
		t.Setenv("AWS_REGION", "us-west-2")

		cfg, err := loadDefaultS3Config()
		if assert.Nil(t, err) {
			assert.Equal(t, "us-west-2", cfg.Region)
			assertS3Credentials(t, cfg, "cache-key", "cache-secret", "cache-token")
		}
	})

	t.Run("profile is used over cache credentials", func(t *testing.T) {
		isolateS3Env(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "cache-token")
		writeSharedCredentials(t, "[cache-profile]\naws_access_key_id = profile-key\naws_secret_access_key = profile-secret\n")
		t.Setenv("SEMAPHORE_CACHE_AWS_PROFILE", "cache-profile")

		cfg, err := loadDefaultS3Config()
		if assert.Nil(t, err) {
			assertS3Credentials(t, cfg, "profile-key", "profile-secret", "")
		}
	})

	t.Run("instance profile is used over cache credentials", func(t *testing.T) {
		isolateS3Env(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "cache-token")
		t.Setenv("SEMAPHORE_CACHE_USE_EC2_INSTANCE_PROFILE", "true")
		t.Setenv("AWS_REGION", "eu-west-1")

		cfg, err := loadDefaultS3Config()
		if assert.Nil(t, err) {
			assert.True(t, aws.IsCredentialsProvider(cfg.Credentials, &ec2rolecreds.Provider{}))
			assert.False(t, aws.IsCredentialsProvider(cfg.Credentials, credentials.StaticCredentialsProvider{}))
		}
	})

	t.Run("default chain without cache credentials", func(t *testing.T) {
		isolateS3Env(t)
		setGenericAWSEnv(t)

		cfg, err := loadDefaultS3Config()
		if assert.Nil(t, err) {
			assert.Equal(t, "us-west-2", cfg.Region)
			assertS3Credentials(t, cfg, "generic-key", "generic-secret", "generic-token")
		}
	})
}

func Test__LoadS3ConfigUsingEndpoint(t *testing.T) {
	t.Run("cache credentials with a session token, region defaults to auto", func(t *testing.T) {
		isolateS3Env(t)
		setGenericAWSEnv(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "cache-token")

		cfg, err := loadS3ConfigUsingEndpoint(testS3URL)
		if assert.Nil(t, err) {
			assert.Equal(t, "auto", cfg.Region)
			assertS3Credentials(t, cfg, "cache-key", "cache-secret", "cache-token")

			endpoint, err := cfg.EndpointResolverWithOptions.ResolveEndpoint(s3.ServiceID, cfg.Region)
			if assert.Nil(t, err) {
				assert.Equal(t, testS3URL, endpoint.URL)
			}
		}
	})

	t.Run("region from SEMAPHORE_CACHE_S3_REGION", func(t *testing.T) {
		isolateS3Env(t)
		setCacheS3Credentials(t, "cache-key", "cache-secret", "")
		t.Setenv("SEMAPHORE_CACHE_S3_REGION", "eu-west-1")

		cfg, err := loadS3ConfigUsingEndpoint(testS3URL)
		if assert.Nil(t, err) {
			assert.Equal(t, "eu-west-1", cfg.Region)
			assertS3Credentials(t, cfg, "cache-key", "cache-secret", "")
		}
	})

	t.Run("default chain without cache credentials", func(t *testing.T) {
		isolateS3Env(t)
		setGenericAWSEnv(t)

		cfg, err := loadS3ConfigUsingEndpoint(testS3URL)
		if assert.Nil(t, err) {
			assert.Equal(t, "auto", cfg.Region)
			assertS3Credentials(t, cfg, "generic-key", "generic-secret", "generic-token")
		}
	})
}

// isolateS3Env clears every variable that selects S3 credentials or a region,
// points the SDK at empty shared files and turns off the instance metadata client.
func isolateS3Env(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		"SEMAPHORE_CACHE_S3_URL",
		"SEMAPHORE_CACHE_S3_KEY",
		"SEMAPHORE_CACHE_S3_SECRET",
		"SEMAPHORE_CACHE_S3_SESSION_TOKEN",
		"SEMAPHORE_CACHE_S3_REGION",
		"SEMAPHORE_CACHE_USE_EC2_INSTANCE_PROFILE",
		"SEMAPHORE_CACHE_AWS_PROFILE",
		"AWS_ACCESS_KEY_ID",
		"AWS_ACCESS_KEY",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SECRET_KEY",
		"AWS_SESSION_TOKEN",
		"AWS_REGION",
		"AWS_DEFAULT_REGION",
		"AWS_PROFILE",
		"AWS_DEFAULT_PROFILE",
		"AWS_ROLE_ARN",
		"AWS_WEB_IDENTITY_TOKEN_FILE",
		"AWS_CONTAINER_CREDENTIALS_FULL_URI",
		"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI",
	} {
		t.Setenv(name, "")
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config")
	credentialsFile := filepath.Join(dir, "credentials")
	assert.Nil(t, os.WriteFile(configFile, nil, 0600))
	assert.Nil(t, os.WriteFile(credentialsFile, nil, 0600))

	t.Setenv("AWS_CONFIG_FILE", configFile)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsFile)
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
}

func setGenericAWSEnv(t *testing.T) {
	t.Helper()
	t.Setenv("AWS_ACCESS_KEY_ID", "generic-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "generic-secret")
	t.Setenv("AWS_SESSION_TOKEN", "generic-token")
	t.Setenv("AWS_REGION", "us-west-2")
}

func setCacheS3Credentials(t *testing.T, key, secret, sessionToken string) {
	t.Helper()
	t.Setenv("SEMAPHORE_CACHE_S3_KEY", key)
	t.Setenv("SEMAPHORE_CACHE_S3_SECRET", secret)
	t.Setenv("SEMAPHORE_CACHE_S3_SESSION_TOKEN", sessionToken)
}

func writeSharedCredentials(t *testing.T, content string) {
	t.Helper()
	assert.Nil(t, os.WriteFile(os.Getenv("AWS_SHARED_CREDENTIALS_FILE"), []byte(content), 0600))
}

func assertS3Credentials(t *testing.T, cfg aws.Config, key, secret, sessionToken string) {
	t.Helper()

	creds, err := cfg.Credentials.Retrieve(context.Background())
	if assert.Nil(t, err) {
		assert.Equal(t, key, creds.AccessKeyID)
		assert.Equal(t, secret, creds.SecretAccessKey)
		assert.Equal(t, sessionToken, creds.SessionToken)
	}
}
