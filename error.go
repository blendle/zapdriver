package zapdriver

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/pkg/errors"
)

type stackdriverFmtError struct{ err error }

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// see https://github.com/googleapis/google-cloud-go/issues/1084#issuecomment-474565019
// this is a hack to get stackdriver to parse the stacktrace
func (e stackdriverFmtError) Error() string {
	if e.err == nil {
		return ""
	}
	stackTrace, ok := errors.Cause(e.err).(stackTracer)
	if !ok {
		stackTrace, ok = e.err.(stackTracer)
	}
	if ok {
		buf := bytes.NewBufferString(e.err.Error())
		// routine id and state aren't available in pure go, so we hard-coded these
		// required for stackdrivers runtime.Stack() format parsing
		buf.WriteString("\n\ngoroutine 1 [running]:")
		for _, frame := range stackTrace.StackTrace() {
			buf.WriteByte('\n')

			pc := uintptr(frame) - 1
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				file, line := fn.FileLine(pc)
				buf.WriteString(fmt.Sprintf("%s()\n\t%s:%d +%#x", fn.Name(), file, line, fn.Entry()))
			}
		}
		return buf.String()
	}
	return e.err.Error()
}
