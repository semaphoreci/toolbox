package validators

import (
	"regexp"
	"strings"

	"github.com/semaphoreci/toolbox/sem-context/pkg/utils"
)

var valueSizeLimit = 20000

func ValidateGetAndDeleteArguments(args []string) error {
	if len(args) != 1 {
		return &utils.Error{ErrorMessage: "Exactly one argument expected", ExitCode: 3}
	}
	return isKeyValid(args[0])
}

func ValidatePutArguments(args []string) error {
	if len(args) != 1 || len(strings.Split(args[0], "=")) != 2 {
		return &utils.Error{ErrorMessage: "Put command expects one argument in form of key=value", ExitCode: 3}
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
	return &utils.Error{
		ErrorMessage: "Key must be between 3 and 256 characters in length, and can contain letters, " +
			"digits, and characters _ and - (no spaces)",
		ExitCode: 3}
}

func isValueValid(value string) error {
	if value == "" {
		return &utils.Error{ErrorMessage: "Value cant be an empty string", ExitCode: 3}
	}
	if len([]byte(value)) > int(valueSizeLimit) {
		return &utils.Error{ErrorMessage: "Value cant be more than 20KB in size", ExitCode: 3}
	}
	return nil
}
