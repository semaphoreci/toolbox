package storage

import "strings"

func (s *SFTPStorage) HasKey(key string) (bool, error) {
	file, err := s.Client.Stat(key)
	if file == nil {
		if err != nil && strings.Contains(err.Error(), "file does not exist") {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
