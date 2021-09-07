package files

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func Unpack(path string) (string, error) {
	restorationPath, err := findRestorationPath(path)
	if err != nil {
		return "", err
	}

	cmd, err := unpackCommand(restorationPath, path)
	if err != nil {
		return "", err
	}

	_, err = cmd.Output()
	if err != nil {
		return "", err
	}

	return restorationPath, nil
}

func unpackCommand(restorationPath, tempFile string) (*exec.Cmd, error) {
	if filepath.IsAbs(restorationPath) {
		return exec.Command("tar", "xzPf", tempFile, "-C", "."), nil
	} else {
		return exec.Command("tar", "xzf", tempFile, "-C", "."), nil
	}
}

func findRestorationPath(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("error opening %s: %v\n", path, err)
		return "", err
	}

	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		fmt.Printf("error creating gzip reader: %v\n", err)
		return "", err
	}

	defer gzipReader.Close()

	tr := tar.NewReader(gzipReader)
	header, err := tr.Next()
	if err == io.EOF {
		fmt.Printf("No files in archive.\n")
		return "", nil
	}

	if err != nil {
		fmt.Printf("Error reading %s: %v\n", path, err)
		return "", err
	}

	return header.Name, nil
}
