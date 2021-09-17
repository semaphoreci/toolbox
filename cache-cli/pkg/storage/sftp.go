package storage

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPStorage struct {
	SSHClient     *ssh.Client
	SFTPClient    *sftp.Client
	StorageConfig StorageConfig
}

type SFTPStorageOptions struct {
	URL            string
	Username       string
	PrivateKeyPath string
	Config         StorageConfig
}

func NewSFTPStorage(options SFTPStorageOptions) (*SFTPStorage, error) {
	sshClient, err := createSSHClient(options)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		fmt.Printf("Error creating sftp client: %v\n", err)
		sshClient.Close()
		return nil, err
	}

	storage := SFTPStorage{
		SSHClient:     sshClient,
		SFTPClient:    sftpClient,
		StorageConfig: options.Config,
	}

	return &storage, nil
}

func (s *SFTPStorage) Config() StorageConfig {
	return s.StorageConfig
}

func createSSHClient(options SFTPStorageOptions) (*ssh.Client, error) {
	pk, _ := ioutil.ReadFile(options.PrivateKeyPath)
	signer, err := ssh.ParsePrivateKey(pk)
	if err != nil {
		fmt.Printf("Error parsing private key: %v\n", err)
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: options.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", options.URL, config)
	if err != nil {
		fmt.Printf("Error dialing ssh: %v\n", err)
		return nil, err
	}

	return sshClient, nil
}
