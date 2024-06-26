package storage

import "context"

func (s *SFTPStorage) Clear(ctx context.Context) error {
	keys, err := s.List(ctx)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		err := s.SFTPClient.Remove(key.Name)
		if err != nil {
			return err
		}
	}

	return nil
}
