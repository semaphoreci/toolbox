package storage

import (
	"fmt"
	"os"
	"time"
)

type Storage interface {
	List() ([]CacheKey, error)
	HasKey(key string) (bool, error)
	Store(key, path string) error
	Delete(key string) error
	Clear() error
	Usage() (*UsageSummary, error)
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
		return NewS3Storage()
	default:
		return nil, fmt.Errorf("cache backend %s is not available", backend)
	}
}
