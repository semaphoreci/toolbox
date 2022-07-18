package utils

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
)

// If error is present, exit and log error message, if it isn't, dont do anything.
// If flag --ignore-failure is set, then exit with code 0, else exit with given error code
func CheckError(err error, code int, msg_for_logging string) {
	if err != nil {
		log.Errorf(msg_for_logging)
		if !flags.IgnoreFailure {
			os.Exit(code)
		} else {
			os.Exit(0)
		}
	}
}
