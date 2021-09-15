package storage

func (s *SFTPStorage) Usage() (*UsageSummary, error) {
	files, err := s.Client.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var totalUsed int64 = 0
	for _, file := range files {
		totalUsed = totalUsed + file.Size()
	}

	return &UsageSummary{
		Used: totalUsed,
		Free: SFTPStorageLimit - totalUsed,
	}, nil
}
