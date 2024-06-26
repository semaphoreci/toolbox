package storage

import "context"

func (s *S3Storage) IsNotEmpty(ctx context.Context) (bool, error) {
	keys, err := s.List(ctx)
	if err != nil {
		return false, err
	}

	return len(keys) != 0, nil
}
