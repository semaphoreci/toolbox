package validators

import (
	"fmt"
	"regexp"
)

var valueSizeLimit = 20000

func IsKeyValid(key string) error {
	keyRegex := regexp.MustCompile(`[A-Za-z0-9-_]{3,256}`)
	if keyRegex.MatchString(key) {
		return nil
	}
	return fmt.Errorf(
		"Key must be between 3 and 256 characters in length, and can contain letters, " +
			"digits, and characters _ and - (no spaces)",
	)
}

func IsValueValid(value string) error {
	if value == "" {
		return fmt.Errorf("Value cant be an empty string")
	}
	if len([]byte(value)) > int(valueSizeLimit) {
		return fmt.Errorf("Value cant be more than 20KB in size")
	}
	return nil
}
