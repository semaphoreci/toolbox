package utils

import (
	"os"
)

// If error exists quits application with given code
func CheckError(err error, code int) {
	if err != nil {
		os.Exit(code)
	}
}
