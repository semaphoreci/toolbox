package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func (s *GCSStorage) Store(ctx context.Context, key, path string) error {
	// #nosec
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	destination := fmt.Sprintf("%s/%s", s.Project, key)
	writer := s.Bucket.Object(destination).NewWriter(ctx)

	_, err = io.Copy(writer, file)
	if err != nil {
		log.Errorf("Error uploading: %v", err)
		_ = file.Close()

		// canceled context will abort the save, closing writer would save a partial object
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return file.Close()
}
