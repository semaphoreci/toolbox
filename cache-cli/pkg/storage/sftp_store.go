package storage

import (
	"os"
)

func (s *SFTPStorage) Store(key, path string) error {
	localFile, err := os.Open(path)
	if err != nil {
		return err
	}

	defer localFile.Close()

	remoteFile, err := s.Client.Create(key)
	if err != nil {
		return err
	}

	defer remoteFile.Close()

	_, err = remoteFile.ReadFrom(localFile)

	return err
}
