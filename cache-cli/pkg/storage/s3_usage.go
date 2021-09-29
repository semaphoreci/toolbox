package storage

func (s *S3Storage) Usage() (*UsageSummary, error) {
	keys, err := s.List()
	if err != nil {
		return nil, err
	}

	var total int64
	for _, key := range keys {
		total = total + key.Size
	}

	return &UsageSummary{
		Used: total,
		Free: -1,
	}, nil
}
