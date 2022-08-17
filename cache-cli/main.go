package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/semaphoreci/toolbox/cache-cli/cmd"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/logging"
	log "github.com/sirupsen/logrus"
)

func main() {
	logfile := OpenLogfile()
	log.SetOutput(logfile)
	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	cmd.Execute()
}

func OpenLogfile() io.Writer {
	// #nosec
	filePath := filepath.Join(os.TempDir(), "cache_log")
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		log.Fatal(err)
	}

	return io.MultiWriter(f, os.Stdout)
}
