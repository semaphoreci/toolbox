package storage

import (
	"context"
	"strings"
)

func (s *SFTPStorage) Delete(ctx context.Context, key string) error {
	err := s.SFTPClient.Remove(key)
	if err != nil && strings.Contains(err.Error(), "file does not exist") {
		return nil
	}

	return err
}
