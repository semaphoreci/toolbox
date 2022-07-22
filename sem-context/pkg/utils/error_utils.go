package utils

import (
	"fmt"
	"os"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
)

// If error is present, exit and log error message, if it isn't, dont do anything.
// If flag --ignore-failure is set, then exit with code 0, else exit with given error code
func CheckError(err error, code int, optional_err_msg ...string) {
	if err != nil {

		err_msg := err.Error()
		for _, optional_msg := range optional_err_msg {
			err_msg = optional_msg
		}

		fmt.Fprintf(os.Stderr, err_msg)
		if !flags.IgnoreFailure {
			os.Exit(code)
		} else {
			os.Exit(0)
		}
	}
}
