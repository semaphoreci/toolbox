package storage

import (
	"io"
)

func (s *SFTPStorage) Restore(key string, writer io.Writer) (int64, error) {
	remoteFile, err := s.SFTPClient.Open(key)
	if err != nil {
		return 0, err
	}

	written, err := remoteFile.WriteTo(writer)
	if err != nil {
		return written, err
	}

	return written, remoteFile.Close()
}
