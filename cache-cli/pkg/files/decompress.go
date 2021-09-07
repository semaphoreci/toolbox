package files

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func Decompress(path string) error {
	cmd, err := decompressionCommand(path)
	if err != nil {
		return err
	}

	_, err = cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func decompressionCommand(path string) (*exec.Cmd, error) {
	restorationPath, err := findRestorationPath(path)
	if err != nil {
		return nil, err
	}

	if filepath.IsAbs(restorationPath) {
		return exec.Command("tar", "xzPf", path, "-C", "."), nil
	} else {
		return exec.Command("tar", "xzf", path, "-C", "."), nil
	}
}

func findRestorationPath(path string) (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command(fmt.Sprintf("tar -ztvf %s 2>/dev/null | head -1 | awk '{print $9}'", path))
	case "linux":
		cmd = exec.Command(fmt.Sprintf("tar -ztvf %s 2>/dev/null | head -1 | awk '{print $6}'", path))
	}

	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	fmt.Printf("Restoration path is %s\n", string(output))
	return string(output), nil
}
