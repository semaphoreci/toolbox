package storage

func (s *SFTPStorage) List() ([]CacheKey, error) {
	files, err := s.Client.ReadDir(".")
	if err != nil {
		return nil, err
	}

	keys := []CacheKey{}
	for _, file := range files {
		storedAt := file.ModTime()
		keys = append(keys, CacheKey{
			Name:     file.Name(),
			Size:     file.Size(),
			StoredAt: &storedAt,
		})
	}

	return keys, nil
}
