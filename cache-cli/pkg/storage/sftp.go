package storage

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPStorage struct {
	Client        *sftp.Client
	StorageConfig StorageConfig
}

type SFTPStorageOptions struct {
	URL            string
	Username       string
	PrivateKeyPath string
	Config         StorageConfig
}

func NewSFTPStorage(options SFTPStorageOptions) (*SFTPStorage, error) {
	client, err := connect(options.URL, options.Username, options.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	storage := SFTPStorage{
		Client:        client,
		StorageConfig: options.Config,
	}

	return &storage, nil
}

func (s *SFTPStorage) Config() StorageConfig {
	return s.StorageConfig
}

func connect(url, username, privateKeyPath string) (*sftp.Client, error) {
	pk, _ := ioutil.ReadFile(privateKeyPath)
	signer, err := ssh.ParsePrivateKey(pk)
	if err != nil {
		fmt.Printf("Error parsing private key: %v\n", err)
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", url, config)
	if err != nil {
		fmt.Printf("Error dialing ssh: %v\n", err)
		return nil, err
	}

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		fmt.Printf("Error creating sftp client: %v\n", err)
		sshClient.Close()
		return nil, err
	}

	return client, nil
}
