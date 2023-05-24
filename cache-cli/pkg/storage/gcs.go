package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

type GCSStorage struct {
	Client        *storage.Client
	Bucket        *storage.BucketHandle
	Project       string
	StorageConfig StorageConfig
}

type GCSStorageOptions struct {
	Bucket  string
	Config  StorageConfig
}

func NewGCSStorage(options GCSStorageOptions) (*GCSStorage, error) {
	return createDefaultGCSStorage(options.Bucket, options.Config)
}

func createDefaultGCSStorage(gcsBucket string, storageConfig StorageConfig) (*GCSStorage, error) {
	client, err := storage.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}

	return &GCSStorage{
		Client:        client,
		Bucket:        client.Bucket(gcsBucket),
		StorageConfig: storageConfig,
	}, nil
}

func (s *GCSStorage) Config() StorageConfig {
	return s.StorageConfig
}
