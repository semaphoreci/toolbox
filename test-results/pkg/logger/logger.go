package logger

import (
	log "github.com/sirupsen/logrus"
)

type logFormatter struct {
}

func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	return append([]byte("* "), entry.Message...), nil
}

// LogEntry ...
var LogEntry = new()

// Level ...
type Level uint32

// Fields ...
type Fields map[string]interface{}

const (
	//InfoLevel ...
	InfoLevel = Level(log.InfoLevel)

	// ErrorLevel ...
	ErrorLevel = Level(log.ErrorLevel)

	// WarnLevel ...
	WarnLevel = Level(log.WarnLevel)

	// DebugLevel ...
	DebugLevel = Level(log.DebugLevel)

	// TraceLevel ...
	TraceLevel = Level(log.TraceLevel)
)

// LogEntry  ...
type logEntry struct {
	logger *log.Logger
}

// new ...
func new() *logEntry {
	logger := log.New()
	logEntry := &logEntry{logger: logger}
	logEntry.logger.SetFormatter(&logFormatter{})
	return logEntry
}

// GetLogger ...
func GetLogger() *log.Logger {
	return LogEntry.logger
}

// SetLogger ...
func SetLogger(logger *log.Logger) {
	LogEntry.logger = logger
	LogEntry.logger.SetFormatter(&logFormatter{})
}

// SetLevel ...
func SetLevel(level Level) {
	LogEntry.logger.SetLevel(log.Level(level))
}

// Error ...
func Error(s string, args ...interface{}) {
	Log(ErrorLevel, s, args...)
}

// Warn ...
func Warn(s string, args ...interface{}) {
	Log(WarnLevel, s, args...)
}

// Info ...
func Info(s string, args ...interface{}) {
	Log(InfoLevel, s, args...)
}

// Debug ...
func Debug(s string, args ...interface{}) {
	Log(DebugLevel, s, args...)
}

// Trace ...
func Trace(s string, args ...interface{}) {
	Log(TraceLevel, s, args...)
}

// Inspect ...
func Inspect(i interface{}) {
	Log(InfoLevel, "=========================================")
	Log(InfoLevel, "%+v", i)
	Log(InfoLevel, "=========================================")
}

// Log ...
func Log(level Level, s string, args ...interface{}) {
	switch level {
	case ErrorLevel:
		LogEntry.logger.Errorf(s+"\n", args...)
	case WarnLevel:
		LogEntry.logger.Warnf(s+"\n", args...)
	case InfoLevel:
		LogEntry.logger.Infof(s+"\n", args...)
	case DebugLevel:
		LogEntry.logger.Debugf(s+"\n", args...)
	case TraceLevel:
		LogEntry.logger.Tracef(s+"\n", args...)
	}
}
