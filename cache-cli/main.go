package main

import (
	"context"
	"io"
	"os"
	"os/signal"
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	cmd.RootCmd.ExecuteContext(ctx)
}

func OpenLogfile() io.Writer {
	// #nosec
	filePath := filepath.Join(os.TempDir(), "cache_log")

	// #nosec
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	/*
	 * We shouldn't fail if we can't create
	 * the log file for whatever reason. Just proceed logging only to stdout.
	 */
	if err != nil {
		log.Errorf("Error creating file '%s': %v - proceeding", filePath, err)
		return os.Stdout
	}

	return io.MultiWriter(f, os.Stdout)
}
