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
	tempFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-*", key))
	if err != nil {
		return nil, err
	}

	bucketKey := fmt.Sprintf("%s/%s", s.Project, key)
	downloader := manager.NewDownloader(s.Client)
	_, err = downloader.Download(context.TODO(), tempFile, &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &bucketKey,
	})

	if err != nil {
		_ = tempFile.Close()
		return nil, err
	}

	return tempFile, tempFile.Close()
}
