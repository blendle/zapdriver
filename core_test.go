package zapdriver

import (
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWithLabels(t *testing.T) {
	fields := []zap.Field{
		zap.String("hello", "world"),
		Label("one", "value"),
		Label("two", "value"),
	}

	labels := newLabels()
	labels.store = map[string]string{"one": "value", "two": "value"}

	want := []zap.Field{
		zap.String("hello", "world"),
		zap.Object("logging.googleapis.com/labels", labels),
	}

	assert.Equal(t, want, (&core{}).withLabels(fields))
}

func TestExtractLabels(t *testing.T) {
	var lbls *labels
	c := &core{
		Core:       zapcore.NewNopCore(),
		permLabels: newLabels(),
		tempLabels: newLabels(),
	}

	fields := []zap.Field{
		zap.String("hello", "world"),
		Label("one", "world"),
		Label("two", "worlds"),
	}

	lbls, fields = c.extractLabels(fields)

	require.Len(t, lbls.store, 2)

	lbls.mutex.RLock()
	assert.Equal(t, "world", lbls.store["one"])
	assert.Equal(t, "worlds", lbls.store["two"])
	lbls.mutex.RUnlock()

	require.Len(t, fields, 1)
	assert.Equal(t, zap.String("hello", "world"), fields[0])
}

func TestWithSourceLocation(t *testing.T) {
	fields := []zap.Field{zap.String("hello", "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}

	want := []zap.Field{
		zap.String("hello", "world"),
		zap.Object(sourceKey, newSource(pc, file, line, ok)),
	}

	assert.Equal(t, want, (&core{}).withSourceLocation(ent, fields))
}

func TestWithSourceLocation_DoesNotOverwrite(t *testing.T) {
	fields := []zap.Field{zap.String(sourceKey, "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}

	want := []zap.Field{
		zap.String(sourceKey, "world"),
	}

	assert.Equal(t, want, (&core{}).withSourceLocation(ent, fields))
}

func TestWithSourceLocation_OnlyWhenDefined(t *testing.T) {
	fields := []zap.Field{zap.String("hello", "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}
	ent.Caller.Defined = false

	want := []zap.Field{
		zap.String("hello", "world"),
	}

	assert.Equal(t, want, (&core{}).withSourceLocation(ent, fields))
}

func TestWithErrorReport(t *testing.T) {
	fields := []zap.Field{zap.String("hello", "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}

	want := []zap.Field{
		zap.String("hello", "world"),
		zap.Object(contextKey, newReportContext(pc, file, line, ok)),
	}

	assert.Equal(t, want, (&core{}).withErrorReport(ent, fields))
}

func TestWithErrorReport_DoesNotOverwrite(t *testing.T) {
	fields := []zap.Field{zap.String(contextKey, "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}

	want := []zap.Field{
		zap.String(contextKey, "world"),
	}

	assert.Equal(t, want, (&core{}).withErrorReport(ent, fields))
}

func TestWithErrorReport_OnlyWhenDefined(t *testing.T) {
	fields := []zap.Field{zap.String("hello", "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}
	ent.Caller.Defined = false

	want := []zap.Field{
		zap.String("hello", "world"),
	}

	assert.Equal(t, want, (&core{}).withErrorReport(ent, fields))
}

func TestWithServiceContext(t *testing.T) {
	fields := []zap.Field{zap.String("hello", "world")}

	want := []zap.Field{
		zap.String("hello", "world"),
		zap.Object(serviceContextKey, newServiceContext("test service", "test version")),
	}

	assert.Equal(t, want, (&core{}).withServiceContext("test service", "test version", fields))
}

func TestWithServiceContext_DoesNotOverwrite(t *testing.T) {
	fields := []zap.Field{zap.String(serviceContextKey, "world")}

	want := []zap.Field{
		zap.String(serviceContextKey, "world"),
	}

	assert.Equal(t, want, (&core{}).withServiceContext("test service", "test version", fields))
}

func TestWrite(t *testing.T) {
	temp := newLabels()
	temp.store = map[string]string{"one": "1", "two": "2"}

	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := &core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: temp,
	}

	fields := []zap.Field{
		zap.String("hello", "world"),
		Label("one", "value"),
		Label("two", "value"),
	}

	err := core.Write(zapcore.Entry{}, fields)
	require.NoError(t, err)

	assert.NotNil(t, logs.All()[0].ContextMap()[labelsKey])
}

// ref: https://github.com/blendle/zapdriver/issues/29
func TestWriteDuplicateLabels(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := &core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: newLabels(),
	}

	fields := []zap.Field{
		Labels(
			Label("hello", "world"),
			Label("hi", "universe"),
		),
	}

	err := core.Write(zapcore.Entry{}, fields)
	require.NoError(t, err)

	assert.Len(t, logs.All()[0].Context, 1)
}

func TestWriteConcurrent(t *testing.T) {
	temp := newLabels()
	temp.store = map[string]string{"one": "1", "two": "2"}
	goRoutines := 8
	counter := int32(10000)

	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := &core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: temp,
	}

	fields := []zap.Field{
		zap.String("hello", "world"),
		Label("one", "value"),
		Label("two", "value"),
	}

	var wg sync.WaitGroup
	wg.Add(goRoutines)
	for i := 0; i < goRoutines; i++ {
		go func() {
			defer wg.Done()
			for atomic.AddInt32(&counter, -1) > 0 {
				err := core.Write(zapcore.Entry{}, fields)
				require.NoError(t, err)
			}
		}()
	}
	wg.Wait()

	assert.NotNil(t, logs.All()[0].ContextMap()[labelsKey])
}

func TestWithAndWrite(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: newLabels(),
	})

	core = core.With([]zapcore.Field{Label("one", "world")})
	err := core.Write(zapcore.Entry{}, []zapcore.Field{Label("two", "worlds")})
	require.NoError(t, err)

	labels := logs.All()[0].ContextMap()[labelsKey].(map[string]interface{})

	assert.Equal(t, "world", labels["one"])
	assert.Equal(t, "worlds", labels["two"])
}

func TestWithAndWrite_MultipleEntries(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: newLabels(),
	})

	core = core.With([]zapcore.Field{Label("one", "world")})
	err := core.Write(zapcore.Entry{}, []zapcore.Field{Label("two", "worlds")})
	require.NoError(t, err)

	labels := logs.All()[0].ContextMap()[labelsKey].(map[string]interface{})
	require.Len(t, labels, 2)

	assert.Equal(t, "world", labels["one"])
	assert.Equal(t, "worlds", labels["two"])

	err = core.Write(zapcore.Entry{}, []zapcore.Field{Label("three", "worlds")})
	require.NoError(t, err)

	labels = logs.All()[1].ContextMap()[labelsKey].(map[string]interface{})
	require.Len(t, labels, 2)

	assert.Equal(t, "world", labels["one"])
	assert.Equal(t, "worlds", labels["three"])
}

func TestWriteReportAllErrors(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: newLabels(),
		config: driverConfig{
			ReportAllErrors: true,
		},
	})

	pc, file, line, ok := runtime.Caller(0)
	// core.With should return with correct config
	core = core.With([]zapcore.Field{Label("one", "world")})
	err := core.Write(zapcore.Entry{
		Level:  zapcore.ErrorLevel,
		Caller: zapcore.NewEntryCaller(pc, file, line, ok),
	}, []zapcore.Field{Label("two", "worlds")})
	require.NoError(t, err)

	context := logs.All()[0].ContextMap()[contextKey].(map[string]interface{})
	rLocation := context["reportLocation"].(map[string]interface{})
	assert.Contains(t, rLocation["filePath"], "zapdriver/core_test.go")
	assert.Equal(t, strconv.Itoa(line), rLocation["lineNumber"])
	assert.Contains(t, rLocation["functionName"], "zapdriver.TestWriteReportAllErrors")

	// Assert that a service context was attached even though service name was not set
	serviceContext := logs.All()[0].ContextMap()[serviceContextKey].(map[string]interface{})
	assert.Equal(t, "unknown", serviceContext["service"])
}

func TestWriteServiceContext(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: newLabels(),
		config: driverConfig{
			ServiceName:    "test service",
			ServiceVersion: "v0.0.1",
		},
	})

	err := core.Write(zapcore.Entry{}, []zapcore.Field{})
	require.NoError(t, err)

	// Assert that a service context was attached even though service name was not set
	serviceContext := logs.All()[0].ContextMap()[serviceContextKey].(map[string]interface{})
	assert.Equal(t, "test service", serviceContext["service"])
	assert.Equal(t, "v0.0.1", serviceContext["version"])
}

func TestWriteReportAllErrors_WithServiceContext(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: newLabels(),
		config: driverConfig{
			ReportAllErrors: true,
			ServiceName:     "test service",
			ServiceVersion:  "v0.0.1",
		},
	})

	pc, file, line, ok := runtime.Caller(0)
	err := core.Write(zapcore.Entry{
		Level:  zapcore.ErrorLevel,
		Caller: zapcore.NewEntryCaller(pc, file, line, ok),
	}, []zapcore.Field{})
	require.NoError(t, err)

	assert.Contains(t, logs.All()[0].ContextMap(), contextKey)

	// Assert that a service context was attached even though service name was not set
	serviceContext := logs.All()[0].ContextMap()[serviceContextKey].(map[string]interface{})
	assert.Equal(t, "test service", serviceContext["service"])
	assert.Equal(t, "v0.0.1", serviceContext["version"])
}

func TestWriteReportAllErrors_InfoLog(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{
		Core:       debugcore,
		permLabels: newLabels(),
		tempLabels: newLabels(),
		config: driverConfig{
			ReportAllErrors: true,
		},
	})

	pc, file, line, ok := runtime.Caller(0)
	err := core.Write(zapcore.Entry{
		Level:  zapcore.InfoLevel,
		Caller: zapcore.NewEntryCaller(pc, file, line, ok),
	}, []zapcore.Field{})
	require.NoError(t, err)

	assert.NotContains(t, logs.All()[0].ContextMap(), contextKey)
	assert.NotContains(t, logs.All()[0].ContextMap(), serviceContextKey)
}

func TestAllLabels(t *testing.T) {
	perm := newLabels()
	perm.store = map[string]string{"one": "1", "two": "2", "three": "3"}

	temp := newLabels()
	temp.store = map[string]string{"one": "ONE", "three": "THREE"}

	core := &core{
		Core:       zapcore.NewNopCore(),
		permLabels: perm,
		tempLabels: temp,
	}

	out := core.allLabels()
	require.Len(t, out.store, 3)

	out.mutex.RLock()
	assert.Equal(t, out.store["one"], "ONE")
	assert.Equal(t, out.store["two"], "2")
	assert.Equal(t, out.store["three"], "THREE")
	out.mutex.RUnlock()
}
