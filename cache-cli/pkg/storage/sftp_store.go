package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func (s *SFTPStorage) Store(ctx context.Context, key, path string) error {
	epochNanos := time.Now().UnixNano()
	tmpKey := fmt.Sprintf("%s-%d", os.Getenv("SEMAPHORE_JOB_ID"), epochNanos)

	localFileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	err = s.allocateSpace(ctx, localFileInfo.Size())
	if err != nil {
		return err
	}

	// #nosec
	localFile, err := os.Open(path)
	if err != nil {
		return err
	}

	remoteTmpFile, err := s.SFTPClient.Create(tmpKey)
	if err != nil {
		_ = localFile.Close()
		return err
	}

	_, err = remoteTmpFile.ReadFrom(localFile)

	if err != nil {
		if rmErr := s.SFTPClient.Remove(tmpKey); rmErr != nil {
			log.Errorf("Error removing temporary file %s: %v", tmpKey, rmErr)
		}

		_ = localFile.Close()
		_ = remoteTmpFile.Close()
		return err
	}

	err = s.SFTPClient.PosixRename(tmpKey, key)
	if err != nil {
		if rmErr := s.SFTPClient.Remove(tmpKey); rmErr != nil {
			log.Errorf("Error removing temporary file %s: %v", tmpKey, rmErr)
		}

		_ = localFile.Close()
		_ = remoteTmpFile.Close()
		return err
	}

	err = remoteTmpFile.Close()
	if err != nil {
		_ = localFile.Close()
		return err
	}

	return localFile.Close()
}

func (s *SFTPStorage) allocateSpace(ctx context.Context, space int64) error {
	usage, err := s.Usage(ctx)
	if err != nil {
		return err
	}

	freeSpace := usage.Free
	if freeSpace < space {
		fmt.Printf("Not enough space, deleting keys based on %s...\n", s.Config().SortKeysBy)
		keys, err := s.List(ctx)
		if err != nil {
			return err
		}

		for freeSpace < space {
			lastKey := keys[len(keys)-1]
			err = s.Delete(ctx, lastKey.Name)
			if err != nil {
				return err
			}

			log.Infof("Key '%s' is deleted.", lastKey.Name)
			freeSpace = freeSpace + lastKey.Size
			keys = keys[:len(keys)-1]
		}
	}

	return nil
}
