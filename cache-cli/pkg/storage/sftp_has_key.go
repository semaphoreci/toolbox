package storage

import (
	"context"
	"strings"
)

func (s *SFTPStorage) HasKey(ctx context.Context, key string) (bool, error) {
	file, err := s.SFTPClient.Stat(key)
	if file == nil {
		if err != nil && strings.Contains(err.Error(), "file does not exist") {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
