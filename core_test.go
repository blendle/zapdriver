package zapdriver

import (
	"runtime"
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
		zap.Object("labels", labels),
	}

	assert.Equal(t, want, (&core{}).withLabels(fields))
}

func TestExtractLabels(t *testing.T) {
	var lbls *labels
	c := &core{zapcore.NewNopCore(), newLabels(), newLabels()}

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

func TestWrite(t *testing.T) {
	temp := newLabels()
	temp.store = map[string]string{"one": "1", "two": "2"}

	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := &core{debugcore, newLabels(), temp}

	fields := []zap.Field{
		zap.String("hello", "world"),
		Label("one", "value"),
		Label("two", "value"),
	}

	err := core.Write(zapcore.Entry{}, fields)
	require.NoError(t, err)

	assert.NotNil(t, logs.All()[0].ContextMap()["labels"])
}

func TestWriteConcurrent(t *testing.T) {
	temp := newLabels()
	temp.store = map[string]string{"one": "1", "two": "2"}
	goRoutines := 8
	counter := int32(10000)

	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := &core{debugcore, newLabels(), temp}

	fields := []zap.Field{
		zap.String("hello", "world"),
		Label("one", "value"),
		Label("two", "value"),
	}

	for i := 0; i < goRoutines; i++ {
		go func() {
			for atomic.AddInt32(&counter, -1) > 0 {
				err := core.Write(zapcore.Entry{}, fields)
				require.NoError(t, err)
			}
		}()
	}

	assert.NotNil(t, logs.All()[0].ContextMap()["labels"])
}

func TestWithAndWrite(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{debugcore, newLabels(), newLabels()})

	core = core.With([]zapcore.Field{Label("one", "world")})
	err := core.Write(zapcore.Entry{}, []zapcore.Field{Label("two", "worlds")})
	require.NoError(t, err)

	labels := logs.All()[0].ContextMap()["labels"].(map[string]interface{})

	assert.Equal(t, "world", labels["one"])
	assert.Equal(t, "worlds", labels["two"])
}

func TestWithAndWrite_MultipleEntries(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{debugcore, newLabels(), newLabels()})

	core = core.With([]zapcore.Field{Label("one", "world")})
	err := core.Write(zapcore.Entry{}, []zapcore.Field{Label("two", "worlds")})
	require.NoError(t, err)

	labels := logs.All()[0].ContextMap()["labels"].(map[string]interface{})
	require.Len(t, labels, 2)

	assert.Equal(t, "world", labels["one"])
	assert.Equal(t, "worlds", labels["two"])

	err = core.Write(zapcore.Entry{}, []zapcore.Field{Label("three", "worlds")})
	require.NoError(t, err)

	labels = logs.All()[1].ContextMap()["labels"].(map[string]interface{})
	require.Len(t, labels, 2)

	assert.Equal(t, "world", labels["one"])
	assert.Equal(t, "worlds", labels["three"])
}

func TestAllLabels(t *testing.T) {
	perm := newLabels()
	perm.store = map[string]string{"one": "1", "two": "2", "three": "3"}

	temp := newLabels()
	temp.store = map[string]string{"one": "ONE", "three": "THREE"}

	core := &core{zapcore.NewNopCore(), perm, temp}

	out := core.allLabels()
	require.Len(t, out.store, 3)

	out.mutex.RLock()
	assert.Equal(t, out.store["one"], "ONE")
	assert.Equal(t, out.store["two"], "2")
	assert.Equal(t, out.store["three"], "THREE")
	out.mutex.RUnlock()
}
