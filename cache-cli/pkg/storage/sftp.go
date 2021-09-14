package storage

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPStorage struct {
	URL            string
	Username       string
	PrivateKeyPath string
	Client         *sftp.Client
}

func NewSFTPStorage(url, username, privateKeyPath string) (*SFTPStorage, error) {
	storage := SFTPStorage{
		URL:            url,
		Username:       username,
		PrivateKeyPath: privateKeyPath,
	}

	err := storage.Connect()
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func (s *SFTPStorage) Connect() error {
	pk, _ := ioutil.ReadFile(s.PrivateKeyPath)
	signer, err := ssh.ParsePrivateKey(pk)
	if err != nil {
		fmt.Printf("Error parsing private key: %v\n", err)
		return err
	}

	config := &ssh.ClientConfig{
		User: s.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", "sftp-server:22", config)
	if err != nil {
		fmt.Printf("Error dialing ssh: %v\n", err)
		return err
	}

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		fmt.Printf("Error creating sftp client: %v\n", err)
		sshClient.Close()
		return err
	}

	s.Client = client
	return nil
}
