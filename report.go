package zapdriver

import (
	"runtime"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const contextKey = "context"

func ErrorReport(pc uintptr, file string, line int, ok bool) zap.Field {
	return zap.Object(contextKey, newReportContext(pc, file, line, ok))
}

type reportLocation struct {
	File     string `json:"filePath"`
	Line     string `json:"lineNumber"`
	Function string `json:"functionName"`
}

// MarshalLogObject implements zapcore.ObjectMarshaller interface.
func (location reportLocation) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("filePath", location.File)
	enc.AddString("lineNumber", location.Line)
	enc.AddString("functionName", location.Function)

	return nil
}

type reportContext struct {
	ReportLocation reportLocation `json:"reportLocation"`
}

// MarshalLogObject implements zapcore.ObjectMarshaller interface.
func (context reportContext) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddObject("reportLocation", context.ReportLocation)

	return nil
}

func newReportContext(pc uintptr, file string, line int, ok bool) *reportContext {
	if !ok {
		return nil
	}

	var function string
	if fn := runtime.FuncForPC(pc); fn != nil {
		function = fn.Name()
	}

	context := &reportContext{
		ReportLocation: reportLocation{
			File:     file,
			Line:     strconv.Itoa(line),
			Function: function,
		},
	}

	return context
}
