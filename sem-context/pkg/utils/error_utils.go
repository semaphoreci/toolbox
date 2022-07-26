package utils

import (
	"fmt"
	"os"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
)

type Error struct {
	ErrorMessage string
	ExitCode     int
}

func (err *Error) Error() string {
	return err.ErrorMessage
}

// If error is present, exit and log error message, if it isn't, dont do anything.
// If flag --ignore-failure is set, then exit with code 0, else exit with given error code
// Error passed to this function must be of type Error defined inside this module
func CheckError(err error) {
	if err != nil {
		castedError := err.(*Error)
		fmt.Fprintf(os.Stderr, err.Error())
		if !flags.IgnoreFailure {
			os.Exit(castedError.ExitCode)
		} else {
			os.Exit(0)
		}
	}
}
