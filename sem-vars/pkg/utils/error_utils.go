package utils

import (
	"os"

	"github.com/semaphoreci/toolbox/sem-vars/pkg/flags"
)

// If error is present, exit, if it isn't, dont do anything.
// If flag --ignore-failure is set, then exit with code 0, else exit with given error code
func CheckError(err error, code int) {
	if err != nil && !flags.IgnoreFailure {
		os.Exit(code)
	} else if err != nil {
		os.Exit(0)
	}
}
