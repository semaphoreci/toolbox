package storage

import "os"

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

	sshSession.Stderr = os.Stderr
	sshSession.Stdin = os.Stdin
	sshSession.Stdout = os.Stdout

	err = sshSession.Run("bash -c 'ls -A1 | xargs rm -rf'")
	if err != nil {
		return err
	}

	return nil
}
