package validators

import (
	"fmt"
	"regexp"
	"strings"
)

var valueSizeLimit = 20000

func ValidateGetAndDeleteArguments(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Exactly one argument expected")
	}
	return isKeyValid(args[0])
}

func ValidatePutArguments(args []string) error {
	if len(args) != 1 || len(strings.Split(args[0], "=")) != 2 {
		return fmt.Errorf("Put command expects one argument in form of key=value")
	}
	err := isKeyValid(strings.Split(args[0], "=")[0])
	if err == nil {
		err = isValueValid(strings.Split(args[0], "=")[1])
	}
	return err
}

func isKeyValid(key string) error {
	keyRegex := regexp.MustCompile(`[A-Za-z0-9-_]{3,256}`)
	if keyRegex.MatchString(key) {
		return nil
	}
	return fmt.Errorf(
		"Key must be between 3 and 256 characters in length, and can contain letters, " +
			"digits, and characters _ and - (no spaces)",
	)
}

func isValueValid(value string) error {
	if value == "" {
		return fmt.Errorf("Value cant be an empty string")
	}
	if len([]byte(value)) > int(valueSizeLimit) {
		return fmt.Errorf("Value cant be more than 20KB in size")
	}
	return nil
}
