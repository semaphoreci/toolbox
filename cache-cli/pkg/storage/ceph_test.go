package storage

import (
	"context"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	assert "github.com/stretchr/testify/assert"
)

func Test__NewCephStorageValidation(t *testing.T) {
	withEnv(t, map[string]string{
		"SEMAPHORE_CACHE_OIDC_TOKEN": "",
		"SEMAPHORE_CACHE_ROLE_ARN":   "",
	}, func() {
		opts := S3StorageOptions{URL: "https://ceph.example", Bucket: "b", Project: "p"}

		_, err := NewCephStorage(opts)
		assert.ErrorContains(t, err, "SEMAPHORE_CACHE_OIDC_TOKEN")

		os.Setenv("SEMAPHORE_CACHE_OIDC_TOKEN", "token")
		_, err = NewCephStorage(opts)
		assert.ErrorContains(t, err, "SEMAPHORE_CACHE_ROLE_ARN")

		os.Setenv("SEMAPHORE_CACHE_ROLE_ARN", "arn:aws:iam::acc:role/rw")
		_, err = NewCephStorage(S3StorageOptions{URL: "", Bucket: "b", Project: "p"})
		assert.ErrorContains(t, err, "SEMAPHORE_CACHE_S3_URL")

		storage, err := NewCephStorage(opts)
		assert.Nil(t, err)
		assert.NotNil(t, storage)
		assert.Equal(t, "b", storage.Bucket)
		assert.Equal(t, "p", storage.Project)
	})
}

func Test__RemoveS3OperationIDMiddleware(t *testing.T) {
	mw := &removeS3OperationIDMiddleware{}

	req := smithyhttp.NewStackRequest().(*smithyhttp.Request)
	parsed, _ := url.Parse("https://ceph.example/semaphore-cache/key?x-id=GetObject&versionId=7")
	req.URL = parsed

	next := middleware.BuildHandlerFunc(
		func(_ context.Context, in middleware.BuildInput) (middleware.BuildOutput, middleware.Metadata, error) {
			return middleware.BuildOutput{}, middleware.Metadata{}, nil
		},
	)

	_, _, err := mw.HandleBuild(context.TODO(), middleware.BuildInput{Request: req}, next)
	assert.Nil(t, err)

	query := req.URL.Query()
	assert.Equal(t, "", query.Get("x-id"))
	// Other query params are preserved.
	assert.Equal(t, "7", query.Get("versionId"))
}

// Exercises the overridden Store (single PutObject) and Restore (sequential
// 8 MiB ranged GETs) against the S3-compatible test backend (MinIO). Uses an
// object larger than the 8 MiB part size so Restore loops over multiple ranges.
func Test__CephStorageStoreAndRestore(t *testing.T) {
	endpoint := os.Getenv("SEMAPHORE_CACHE_S3_URL")
	if endpoint == "" {
		t.Skip("SEMAPHORE_CACHE_S3_URL not set; skipping Ceph IO test")
	}

	s3s, err := NewS3Storage(S3StorageOptions{
		URL:     endpoint,
		Bucket:  "semaphore-cache",
		Project: "cache-cli",
		Config:  StorageConfig{MaxSpace: math.MaxInt64, SortKeysBy: SortByStoreTime},
	})
	if !assert.Nil(t, err) {
		return
	}

	storage := &CephStorage{S3Storage: s3s}
	_ = storage.Clear()

	content := strings.Repeat("ceph-cache-payload-", 1_200_000) // ~22 MiB -> 3 ranges
	file, _ := ioutil.TempFile(os.TempDir(), "*")
	_, _ = file.WriteString(content)
	_ = file.Close()
	defer os.Remove(file.Name())

	assert.Nil(t, storage.Store("ceph-key", file.Name()))

	keys, err := storage.List()
	assert.Nil(t, err)
	assert.Len(t, keys, 1)

	restored, err := storage.Restore("ceph-key")
	assert.Nil(t, err)
	defer os.Remove(restored.Name())

	got, err := ioutil.ReadFile(restored.Name())
	assert.Nil(t, err)
	assert.Equal(t, content, string(got))
}

func withEnv(t *testing.T, vars map[string]string, fn func()) {
	t.Helper()

	previous := map[string]*string{}
	for k, v := range vars {
		if old, ok := os.LookupEnv(k); ok {
			previous[k] = &old
		} else {
			previous[k] = nil
		}

		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}

	defer func() {
		for k, old := range previous {
			if old == nil {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, *old)
			}
		}
	}()

	fn()
}
