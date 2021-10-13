package utils

import (
	"fmt"
	"os"
)

func Check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())

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
		fmt.Fprintf(os.Stderr, "error: %+v\n", message)

		os.Exit(1)
	}
}
