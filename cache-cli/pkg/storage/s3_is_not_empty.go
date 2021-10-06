package storage

func (s *S3Storage) IsNotEmpty() (bool, error) {
	keys, err := s.List()
	if err != nil {
		return false, err
	}

	return len(keys) != 0, nil
}
