package zapdriver

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
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
	z, b := testZap()
	z.Error("An error occurred", zap.Error(erroring()))

	fmt.Println(makeStackTraceStable(b.String()))
	// Output: {"severity":"ERROR","caller":"zapdriver/error_test.go:48","message":"An error occurred","exception":"bar: foo","logging.googleapis.com/labels":{},"logging.googleapis.com/sourceLocation":{"file":"/error_test.go","line":"48","function":"github.com/blendle/zapdriver.ExampleUseFmtStackTracing"}}
}

func testZap() (*zap.Logger, *bytes.Buffer) {
	var b = bytes.NewBuffer(nil)
	conf := NewDevelopmentConfig()
	conf.EncoderConfig.TimeKey = ""
	zap.RegisterSink("example", func(u *url.URL) (zap.Sink, error) { return testSink{b}, nil })
	conf.OutputPaths = []string{"example://"} // to generate example output

	z, err := conf.Build(WrapCore(
		FmtStackTraces(true),
		ReportAllErrors(false),
	))

	if err != nil {
		panic(err)
	}
	return z, b
}

func erroring() error {
	return errors.Wrap(fmt.Errorf("foo"), "bar")
}

func makeStackTraceStable(str string) string {
	re := regexp.MustCompile(`(?m)^[\t\\t].+(\/\S+):\d+ \+0x.+$`)
	str = re.ReplaceAllString(str, "\t${1}:42 +0x1337")
	dir, _ := os.Getwd()
	str = strings.ReplaceAll(str, dir, "")
	return str
}

type testSink struct {
	io.Writer
}

func (testSink) Close() error {
	return nil
}

func (testSink) Sync() error {
	return nil
}
