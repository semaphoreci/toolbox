package storage

func (s *SFTPStorage) Delete(key string) error {
	return s.Client.Remove(key)
}
