package storage

import (
	"fmt"
	"io/ioutil"
	"os"
)

func (s *SFTPStorage) Restore(key string) (*os.File, error) {
	localFile, err := ioutil.TempFile("/tmp", fmt.Sprintf("%s-*", key))
	if err != nil {
		return nil, err
	}

	defer localFile.Close()

	remoteFile, err := s.Client.Open(key)
	if err != nil {
		os.Remove(localFile.Name())
		return nil, err
	}

	_, err = localFile.ReadFrom(remoteFile)
	if err != nil {
		return nil, err
	}

	return localFile, nil
}
