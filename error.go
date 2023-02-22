package zapdriver

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

func ErrFormat(err error) string {
	if err == nil {
		return ""
	}
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	cause := errors.Cause(err)
	if stackTrace, ok := cause.(stackTracer); ok {
		buf := bytes.Buffer{}
		for i, frame := range stackTrace.StackTrace() {
			if i == 0 {
				buf.WriteByte('\n')
			}

			buf.WriteString(fmt.Sprintf("\n%+v", frame))
		}
		return err.Error() + buf.String()
	}
	return err.Error()
}

func FormatErrorField(field *zapcore.Field) {
	if err, ok := field.Interface.(error); ok {
		field.Interface = ErrFormat(err)
	}
}
