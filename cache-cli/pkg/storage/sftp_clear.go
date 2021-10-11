package storage

func (s *SFTPStorage) Clear() error {
	keys, err := s.List()
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
