package files

import (
	"crypto/md5" // #nosec
	"encoding/hex"
	"io"
	"os"
)

func GenerateChecksum(filePath string) (string, error) {
	// #nosec
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	// #nosec
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
