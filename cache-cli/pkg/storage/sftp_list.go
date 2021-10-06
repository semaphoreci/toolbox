package storage

import "sort"

func (s *SFTPStorage) List() ([]CacheKey, error) {
	files, err := s.SFTPClient.ReadDir(".")
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

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].StoredAt.After(*keys[j].StoredAt)
	})

	return keys, nil
}
