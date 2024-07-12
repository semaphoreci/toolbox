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
	Project string
	Config  StorageConfig
}

func NewGCSStorage(ctx context.Context, options GCSStorageOptions) (*GCSStorage, error) {
	return createDefaultGCSStorage(ctx, options.Bucket, options.Project, options.Config)
}

func createDefaultGCSStorage(ctx context.Context, gcsBucket string, project string, storageConfig StorageConfig) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCSStorage{
		Client:        client,
		Bucket:        client.Bucket(gcsBucket),
		Project:       project,
		StorageConfig: storageConfig,
	}, nil
}

func (s *GCSStorage) Config() StorageConfig {
	return s.StorageConfig
}
