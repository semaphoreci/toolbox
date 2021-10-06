package storage

func (s *SFTPStorage) IsNotEmpty() (bool, error) {
	keys, err := s.List()
	if err != nil {
		return false, err
	}

	return len(keys) != 0, nil
}
