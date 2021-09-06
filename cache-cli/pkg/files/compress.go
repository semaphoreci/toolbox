package files

import (
	"fmt"
	"os/exec"
	"time"
)

func Compress(key, path string) (string, error) {
	epochNanos := time.Now().Nanosecond()
	temporaryFileName := fmt.Sprintf("/tmp/%s-%d", key, epochNanos)

	cmd := exec.Command("tar", "czPf", temporaryFileName, path)
	_, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return temporaryFileName, nil
}
