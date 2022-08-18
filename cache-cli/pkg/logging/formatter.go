package logging

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type CustomFormatter struct {
}

// We just care about the actual message here
func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
}
