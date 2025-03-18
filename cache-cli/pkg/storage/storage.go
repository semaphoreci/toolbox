package storage

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type Storage interface {
	List() ([]CacheKey, error)
	HasKey(key string) (bool, error)
	Store(key, path string) error
	Restore(key string) (*os.File, error)
	Delete(key string) error
	Clear() error
	Usage() (*UsageSummary, error)
	IsNotEmpty() (bool, error)
	Config() StorageConfig
}

const SortBySize = "SIZE"
const SortByStoreTime = "STORE_TIME"
const SortByAccessTime = "ACCESS_TIME"

var ValidSortByKeys = []string{SortBySize, SortByStoreTime, SortByAccessTime}

type StorageConfig struct {
	MaxSpace   int64
	SortKeysBy string
}

func (c *StorageConfig) Validate() error {
	if contains(c.SortKeysBy, ValidSortByKeys) {
		return nil
	}

	return fmt.Errorf("sorting keys by '%s' is not supported", c.SortKeysBy)
}

type CacheKey struct {
	Name           string
	StoredAt       *time.Time
	LastAccessedAt *time.Time
	Size           int64
}

type UsageSummary struct {
	Free int64
	Used int64
}

func InitStorage() (Storage, error) {
	return InitStorageWithConfig(StorageConfig{SortKeysBy: SortByStoreTime})
}

func InitStorageWithConfig(config StorageConfig) (Storage, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	backend := os.Getenv("SEMAPHORE_CACHE_BACKEND")
	if backend == "" {
		return nil, fmt.Errorf("no SEMAPHORE_CACHE_BACKEND environment variable set")
	}

	switch backend {
	case "s3":
		project := os.Getenv("SEMAPHORE_PROJECT_ID")
		if project == "" {
			return nil, fmt.Errorf("no SEMAPHORE_PROJECT_ID set")
		}

		s3Bucket := os.Getenv("SEMAPHORE_CACHE_S3_BUCKET")
		if s3Bucket == "" {
			return nil, fmt.Errorf("no SEMAPHORE_CACHE_S3_BUCKET set")
		}

		return NewS3Storage(S3StorageOptions{
			URL:     os.Getenv("SEMAPHORE_CACHE_S3_URL"),
			Bucket:  s3Bucket,
			Project: project,
			Config:  StorageConfig{MaxSpace: math.MaxInt64, SortKeysBy: config.SortKeysBy},
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
			Config:         buildStorageConfig(config, 9*1024*1024*1024),
		})
	case "gcs":
		project := os.Getenv("SEMAPHORE_PROJECT_ID")
		if project == "" {
			return nil, fmt.Errorf("no SEMAPHORE_PROJECT_ID set")
		}

		gcsBucket := os.Getenv("SEMAPHORE_CACHE_GCS_BUCKET")
		if gcsBucket == "" {
			return nil, fmt.Errorf("no SEMAPHORE_CACHE_GCS_BUCKET set")
		}

		return NewGCSStorage(GCSStorageOptions{
			Bucket:  gcsBucket,
			Project: project,
			Config:  StorageConfig{MaxSpace: math.MaxInt64, SortKeysBy: config.SortKeysBy},
		})
	default:
		return nil, fmt.Errorf("cache backend '%s' is not available", backend)
	}
}

func buildStorageConfig(config StorageConfig, defaultValue int64) StorageConfig {
	cacheSizeEnvVar := os.Getenv("CACHE_SIZE")
	if cacheSizeEnvVar == "" {
		return StorageConfig{MaxSpace: defaultValue, SortKeysBy: config.SortKeysBy}
	}

	cacheSize, err := strconv.ParseInt(cacheSizeEnvVar, 10, 64)
	if err != nil {
		log.Errorf("Couldn't parse CACHE_SIZE value of '%s' - using default value for storage backend", cacheSizeEnvVar)
		return StorageConfig{MaxSpace: defaultValue, SortKeysBy: config.SortKeysBy}
	}

	// CACHE_SIZE receives kb
	return StorageConfig{MaxSpace: cacheSize * 1024, SortKeysBy: config.SortKeysBy}
}

func contains(item string, items []string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}

	return false
}
