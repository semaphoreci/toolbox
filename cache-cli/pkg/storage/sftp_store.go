package storage

import (
	"fmt"
	"os"
	"time"
)

func (s *SFTPStorage) Store(key, path string) error {
	epochNanos := time.Now().Nanosecond()
	tmpKey := fmt.Sprintf("%s-%d", key, epochNanos)

	localFileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	err = s.allocateSpace(localFileInfo.Size())
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
			fmt.Printf("Error removing temporary file %s: %v\n", tmpKey, rmErr)
		}

		_ = localFile.Close()
		_ = remoteTmpFile.Close()
		return err
	}

	err = s.SFTPClient.PosixRename(tmpKey, key)
	if err != nil {
		if rmErr := s.SFTPClient.Remove(tmpKey); rmErr != nil {
			fmt.Printf("Error removing temporary file %s: %v\n", tmpKey, rmErr)
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

func (s *SFTPStorage) allocateSpace(space int64) error {
	usage, err := s.Usage()
	if err != nil {
		return err
	}

	freeSpace := usage.Free
	if freeSpace < space {
		fmt.Printf("Not enough space, deleting the oldest keys...\n")
		keys, err := s.List()
		if err != nil {
			return err
		}

		for freeSpace < space {
			lastKey := keys[len(keys)-1]
			err = s.Delete(lastKey.Name)
			if err != nil {
				return err
			}

			fmt.Printf("Key '%s' is deleted.\n", lastKey.Name)
			freeSpace = freeSpace + lastKey.Size
			keys = keys[:len(keys)-1]
		}
	}

	return nil
}
