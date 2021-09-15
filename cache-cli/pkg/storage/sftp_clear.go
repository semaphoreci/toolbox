package storage

func (s *SFTPStorage) Clear() error {
	files, err := s.Client.ReadDir(".")
	if err != nil {
		return err
	}

	for _, file := range files {
		err := s.Client.Remove(file.Name())
		if err != nil {
			return err
		}
	}

	return nil
}
