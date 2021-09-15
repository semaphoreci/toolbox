package storage

import "strings"

func (s *SFTPStorage) Delete(key string) error {
	err := s.Client.Remove(key)
	if err != nil && strings.Contains(err.Error(), "file does not exist") {
		return nil
	}

	return err
}
