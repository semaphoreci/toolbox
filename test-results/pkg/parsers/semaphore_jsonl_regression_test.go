package parsers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/stretchr/testify/assert"
)

func semaphoreJSONLWrite(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "report.semaphore.jsonl")
	assert.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestSemaphoreJSONLMixedFrameworkWarning(t *testing.T) {
	p := NewSemaphoreJSONL()
	path := semaphoreJSONLWrite(t, strings.Join([]string{
		`{"kind":"report","schema":"semaphore/test-report","version":"2.0-draft.2","reporter":"r 1","framework":"rspec"}`,
		`{"kind":"test","identity":{"name":"a"},"state":"passed","tags":{},"attempts":[{"index":0,"state":"passed"}]}`,
		`{"kind":"summary","tests":1,"complete":true}`,
		`{"kind":"report","schema":"semaphore/test-report","version":"2.0-draft.2","reporter":"r 1","framework":"exunit"}`,
		`{"kind":"test","identity":{"name":"b"},"state":"passed","tags":{},"attempts":[{"index":0,"state":"passed"}]}`,
		`{"kind":"summary","tests":1,"complete":true}`,
	}, "\n")+"\n")

	results := p.Parse(path)
	assert.Equal(t, parser.StatusSuccess, results.Status)
	assert.Contains(t, results.StatusMessage, "mixed frameworks in stream")
	assert.Contains(t, results.StatusMessage, `"rspec"`)
}

func TestSemaphoreJSONLSameFrameworkConcatNoWarning(t *testing.T) {
	p := NewSemaphoreJSONL()
	path := semaphoreJSONLWrite(t, strings.Join([]string{
		`{"kind":"report","schema":"semaphore/test-report","version":"2.0-draft.2","reporter":"r 1","framework":"rspec"}`,
		`{"kind":"test","identity":{"name":"a"},"state":"passed","tags":{},"attempts":[{"index":0,"state":"passed"}]}`,
		`{"kind":"summary","tests":1,"complete":true}`,
		`{"kind":"report","schema":"semaphore/test-report","version":"2.0-draft.2","reporter":"r 1","framework":"rspec"}`,
		`{"kind":"test","identity":{"name":"b"},"state":"passed","tags":{},"attempts":[{"index":0,"state":"passed"}]}`,
		`{"kind":"summary","tests":1,"complete":true}`,
	}, "\n")+"\n")

	results := p.Parse(path)
	assert.Equal(t, parser.StatusSuccess, results.Status)
	assert.NotContains(t, results.StatusMessage, "mixed frameworks")
}

func TestSemaphoreJSONLMissingHeaderIsError(t *testing.T) {
	p := NewSemaphoreJSONL()
	path := semaphoreJSONLWrite(t, strings.Join([]string{
		`{"kind":"test","identity":{"name":"a"},"state":"passed","tags":{},"attempts":[{"index":0,"state":"passed"}]}`,
		`{"kind":"summary","tests":1,"complete":true}`,
	}, "\n")+"\n")

	results := p.Parse(path)
	assert.Equal(t, parser.StatusError, results.Status)
	assert.Contains(t, results.StatusMessage, "missing report header")
}
