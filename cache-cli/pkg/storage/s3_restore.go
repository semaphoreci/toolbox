package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Storage) Restore(key string) (*os.File, error) {
	tempFile, err := ioutil.TempFile("/tmp", fmt.Sprintf("%s-*", key))
	if err != nil {
		return nil, err
	}

	defer tempFile.Close()

	bucketKey := fmt.Sprintf("%s/%s", s.project, key)
	downloader := manager.NewDownloader(s.client)
	_, err = downloader.Download(context.TODO(), tempFile, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &bucketKey,
	})

	return tempFile, err
}
