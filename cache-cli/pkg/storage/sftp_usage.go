package storage

import "context"

func (s *SFTPStorage) Usage(ctx context.Context) (*UsageSummary, error) {
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
