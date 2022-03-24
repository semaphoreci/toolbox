package storage

import (
	"fmt"
	"io/ioutil"
	"os"
)

func (s *SFTPStorage) Restore(key string) (*os.File, error) {
	localFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-*", key))
	if err != nil {
		return nil, err
	}

	remoteFile, err := s.SFTPClient.Open(key)
	if err != nil {
		_ = localFile.Close()
		_ = os.Remove(localFile.Name())
		return nil, err
	}

	_, err = localFile.ReadFrom(remoteFile)
	if err != nil {
		_ = localFile.Close()
		_ = remoteFile.Close()
		return nil, err
	}

	err = remoteFile.Close()
	if err != nil {
		_ = localFile.Close()
		return nil, err
	}

	return localFile, localFile.Close()
}
