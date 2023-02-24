package zapdriver

import (
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type fakeErr struct{}

// manually set the frames to allow asserting stacktraces
func (fakeErr) StackTrace() errors.StackTrace {
	pc1, _, _, _ := runtime.Caller(0)
	pc2, _, _, _ := runtime.Caller(0)
	return []errors.Frame{
		errors.Frame(pc1),
		errors.Frame(pc2),
	}
}
func (fakeErr) Error() string {
	return "fake error: underlying error"
}

func TestFmtStack(t *testing.T) {
	stacktrace := stackdriverFmtError{fakeErr{}}.Error()
	assert.Equal(t, `fake error: underlying error

goroutine 1 [running]:
github.com/blendle/zapdriver.fakeErr.StackTrace()
	/error_test.go:18 +0x1337
github.com/blendle/zapdriver.fakeErr.StackTrace()
	/error_test.go:19 +0x1337`, makeStackTraceStable(stacktrace))
}

// cleanup local paths & local function pointers
func makeStackTraceStable(str string) string {
	re := regexp.MustCompile(`(?m)^\t.+(\/\S+:\d+) \+0x.+$`)
	str = re.ReplaceAllString(str, "\t${1} +0x1337")
	dir, _ := os.Getwd()
	str = strings.ReplaceAll(str, dir, "")
	return str
}
