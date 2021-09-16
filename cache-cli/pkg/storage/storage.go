package storage

import (
	"fmt"
	"math"
	"os"
	"time"
)

type Storage interface {
	List() ([]CacheKey, error)
	HasKey(key string) (bool, error)
	Store(key, path string) error
	Restore(key string) (*os.File, error)
	Delete(key string) error
	Clear() error
	Usage() (*UsageSummary, error)
	Config() StorageConfig
}

type StorageConfig struct {
	MaxSpace int64
}

type CacheKey struct {
	Name     string
	StoredAt *time.Time
	Size     int64
}

type UsageSummary struct {
	Free int64
	Used int64
}

func InitStorage() (Storage, error) {
	backend := os.Getenv("SEMAPHORE_CACHE_BACKEND")
	if backend == "" {
		return nil, fmt.Errorf("no SEMAPHORE_CACHE_BACKEND environment variable set")
	}

	switch backend {
	case "s3":
		project := os.Getenv("SEMAPHORE_PROJECT_NAME")
		if project == "" {
			return nil, fmt.Errorf("no SEMAPHORE_PROJECT_NAME set")
		}

		s3Bucket := os.Getenv("SEMAPHORE_CACHE_S3_BUCKET")
		if s3Bucket == "" {
			return nil, fmt.Errorf("no SEMAPHORE_CACHE_S3_BUCKET set")
		}

		return NewS3Storage(S3StorageOptions{
			URL:     os.Getenv("SEMAPHORE_CACHE_S3_URL"),
			Bucket:  s3Bucket,
			Project: project,
			Config:  StorageConfig{MaxSpace: math.MaxInt64},
		})
	case "sftp":
		url := os.Getenv("SEMAPHORE_CACHE_URL")
		if url == "" {
			return nil, fmt.Errorf("no SEMAPHORE_CACHE_URL set")
		}

		username := os.Getenv("SEMAPHORE_CACHE_USERNAME")
		if username == "" {
			return nil, fmt.Errorf("no SEMAPHORE_CACHE_USERNAME set")
		}

		privateKeyPath := os.Getenv("SEMAPHORE_CACHE_PRIVATE_KEY_PATH")
		if privateKeyPath == "" {
			return nil, fmt.Errorf("no SEMAPHORE_CACHE_PRIVATE_KEY_PATH set")
		}

		return NewSFTPStorage(SFTPStorageOptions{
			URL:            url,
			Username:       username,
			PrivateKeyPath: privateKeyPath,
			Config:         StorageConfig{MaxSpace: 9 * 1024 * 1024 * 1024},
		})
	default:
		return nil, fmt.Errorf("cache backend '%s' is not available", backend)
	}
}
