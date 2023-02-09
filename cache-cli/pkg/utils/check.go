package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func Check(err error) {
	if err != nil {
		log.Errorf("error: %s", err.Error())

		failOnError := os.Getenv("CACHE_FAIL_ON_ERROR")
		if failOnError == "true" {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}

func CheckWithMessage(err error, message string) {
	if err != nil {
		log.Errorf("error: %+v", message)

		os.Exit(1)
	}
}
