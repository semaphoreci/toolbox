package storage

import (
	"fmt"
	"os"
	"time"
)

type Storage interface {
	List() (*ListResult, error)
	HasKey(key string) (bool, error)
}

type ListResult struct {
	Keys []CacheKey
}

type CacheKey struct {
	Name      string
	UpdatedAt *time.Time
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
