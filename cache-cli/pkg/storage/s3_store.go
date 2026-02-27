package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

func (s *S3Storage) Store(key, path string) error {
	// #nosec
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	destination := fmt.Sprintf("%s/%s", s.Project, key)
	_, err = s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &s.Bucket,
		Key:    &destination,
		Body:   file,
	})

	if err != nil {
		log.Errorf("Error uploading: %v", err)
		return err
	}

	return file.Close()
}
