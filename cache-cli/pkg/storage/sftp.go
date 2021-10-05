package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
	sshKeyPath := resolvePath(options.PrivateKeyPath)
	bytes, err := ioutil.ReadFile(sshKeyPath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", sshKeyPath, err)
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKey(bytes)
	if err != nil {
		fmt.Printf("Error parsing private key: %v\n", err)
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: options.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
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

func resolvePath(path string) string {
	if strings.HasPrefix(path, "~") {
		return strings.Replace(path, "~", os.Getenv("HOME"), 1)
	}

	return path
}
