package storage

func (s *SFTPStorage) Usage() (*UsageSummary, error) {
	files, err := s.SFTPClient.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var totalUsed int64
	for _, file := range files {
		totalUsed = totalUsed + file.Size()
	}

	return &UsageSummary{
		Used: totalUsed,
		Free: s.Config().MaxSpace - totalUsed,
	}, nil
}
