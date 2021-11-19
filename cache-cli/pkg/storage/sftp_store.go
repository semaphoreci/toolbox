package storage

import (
	"fmt"
	"os"
)

func (s *SFTPStorage) Store(key, path string) error {
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

	remoteFile, err := s.SFTPClient.Create(key)
	if err != nil {
		_ = localFile.Close()
		return err
	}

	_, err = remoteFile.ReadFrom(localFile)
	if err != nil {
		_ = localFile.Close()
		_ = remoteFile.Close()
		return err
	}

	err = remoteFile.Close()
	if err != nil {
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
