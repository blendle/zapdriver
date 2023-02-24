package zapdriver

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakeErr struct{}

// manually set the frames to allow asserting stacktraces
func (fakeErr) StackTrace() errors.StackTrace {
	var pcs = make([]uintptr, 2)
	runtime.Callers(3, pcs)

	return []errors.Frame{
		errors.Frame(pcs[0]),
		errors.Frame(pcs[1]),
	}
}
func (fakeErr) Error() string {
	return "fake error: underlying error"
}

func TestFmtStack(t *testing.T) {
	stacktrace := stackdriverFmtError{fakeErr{}}.Error()
	assert.Equal(t, `fake error: underlying error

goroutine 1 [running]:
github.com/blendle/zapdriver.TestFmtStack()
	/error_test.go:42 +0x1337
testing.tRunner()
	/testing.go:42 +0x1337`, makeStackTraceStable(stacktrace))
}

func ExampleUseFmtStackTracing() {
	zap.RegisterSink("example", func(u *url.URL) (zap.Sink, error) {
		return os.Stdout, nil
	})
	conf := NewDevelopmentConfig()
	conf.OutputPaths = []string{"stdout"} // to generate example output
	z, err := conf.Build(WrapCore(
		FmtStackTraces(true),
		ReportAllErrors(true),
		ServiceName("my-service"),
	))

	if err != nil {
		panic(err)
	}

	z.Error("An error occurred", zap.Error(erroring()))
	// Output: foo
}

func erroring() error {
	return errors.Wrap(fmt.Errorf("foo"), "bar")
}

func makeStackTraceStable(str string) string {
	re := regexp.MustCompile(`(?m)^\t.+(\/\S+):\d+ \+0x.+$`)
	return re.ReplaceAllString(str, "\t${1}:42 +0x1337")
}
