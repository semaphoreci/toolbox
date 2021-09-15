package storage

func (s *SFTPStorage) HasKey(key string) (bool, error) {
	file, err := s.Client.Stat(key)
	if file == nil {
		return false, err
	}

	return true, nil
}
