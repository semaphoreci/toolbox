package utils

import (
	"bytes"
	"io"
	"os"
)

type OutputCapturer struct {
	oldStdout *os.File
	reader    *os.File
	writer    *os.File
}

func CreateOutputCapturer() *OutputCapturer {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	return &OutputCapturer{
		oldStdout: old,
		reader:    r,
		writer:    w,
	}
}

func (o *OutputCapturer) Done() string {
	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, o.reader)
		outC <- buf.String()
	}()

	_ = o.writer.Close()
	os.Stdout = o.oldStdout
	output := <-outC

	return output
}
