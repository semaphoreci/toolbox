package storage

func (s *SFTPStorage) Clear() error {
	keys, err := s.List()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	sshSession, err := s.SSHClient.NewSession()
	if err != nil {
		return err
	}

	defer sshSession.Close()

	err = sshSession.Run("rm ./*")
	if err != nil {
		return err
	}

	return nil
}
