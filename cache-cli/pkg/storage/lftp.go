package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var CommandTemplate = `#! env bash
echo -e "%s" | lftp sftp://%s:DUMMY@%s -e "set sftp:auto-confirm yes; set sftp:connect-program \"ssh -a -x -i %s\""
`

type LFTPStorage struct {
	URL            string
	Username       string
	PrivateKeyPath string
}

func NewLFTPStorage(url, username, privateKeyPath string) (*LFTPStorage, error) {
	return &LFTPStorage{
		URL:            url,
		Username:       username,
		PrivateKeyPath: privateKeyPath,
	}, nil
}

func (s *LFTPStorage) ExecuteCommand(command string) (string, error) {
	tempFile, err := ioutil.TempFile("/tmp", "*.sh")
	if err != nil {
		return "", err
	}

	defer func() {
		err := tempFile.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v", err)
		}

		err = os.Remove(tempFile.Name())
		if err != nil {
			fmt.Printf("Error removing file: %v", err)
		}
	}()

	err = os.Chmod(tempFile.Name(), os.ModePerm)
	if err != nil {
		return "", err
	}

	bashCmd := fmt.Sprintf(CommandTemplate, command, s.Username, s.URL, s.PrivateKeyPath)
	tempFile.WriteString(bashCmd)

	cmd := exec.Command("bash", tempFile.Name())
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("%s", err.(*exec.ExitError).Stderr)
	}

	return string(output), err
}
