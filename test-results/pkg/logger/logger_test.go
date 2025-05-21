package logger_test

import (
	"fmt"
	"testing"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func Test_SetLogger(t *testing.T) {
	myLogger, _ := test.NewNullLogger()
	logger.SetLogger(myLogger)

	assert.Equal(t, myLogger, logger.GetLogger(), "Should properly set logger")
}

func Test_GetLogger(t *testing.T) {
	myLogger, _ := test.NewNullLogger()
	logger.SetLogger(myLogger)

	assert.Equal(t, myLogger, logger.GetLogger(), "Should properly get logger")
}

func Test_Debug(t *testing.T) {
	myLogger, hook := test.NewNullLogger()
	logger.SetLogger(myLogger)
	logger.SetLevel(logger.DebugLevel)

	logger.Debug("Debug")
	assert.Equal(t, "Debug\n", hook.LastEntry().Message)
}

func Test_Warn(t *testing.T) {
	myLogger, hook := test.NewNullLogger()
	logger.SetLogger(myLogger)
	logger.SetLevel(logger.WarnLevel)

	logger.Warn("Warn")
	assert.Equal(t, "Warn\n", hook.LastEntry().Message)
}

func Test_Error(t *testing.T) {
	myLogger, hook := test.NewNullLogger()
	logger.SetLogger(myLogger)
	logger.SetLevel(logger.ErrorLevel)

	logger.Error("Error")
	assert.Equal(t, "Error\n", hook.LastEntry().Message)
}

func Test_Info(t *testing.T) {
	myLogger, hook := test.NewNullLogger()
	logger.SetLogger(myLogger)
	logger.SetLevel(logger.InfoLevel)

	logger.Info("Info")
	assert.Equal(t, "Info\n", hook.LastEntry().Message)
}

func Test_Log(t *testing.T) {
	levels := []logger.Level{logger.InfoLevel, logger.WarnLevel, logger.ErrorLevel, logger.DebugLevel, logger.TraceLevel}

	testCases := []struct {
		desc   string
		msg    string
		fields logger.Fields
	}{
		{
			desc: "Works in info Level",
			msg:  "Testing logs ...",
		},
	}

	myLogger, hook := test.NewNullLogger()
	logger.SetLogger(myLogger)

	for _, tC := range testCases {
		for _, level := range levels {
			t.Run(tC.desc+" on level "+fmt.Sprint(level), func(t *testing.T) {
				logger.SetLevel(level)
				logger.Log(level, tC.msg)
				assert.Equal(t, tC.msg+"\n", hook.LastEntry().Message)

				hook.Reset()
			})
		}
	}
}
