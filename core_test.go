package zapdriver

import (
	"runtime"
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

	want := []zap.Field{
		zap.String("hello", "world"),
		zap.Object("labels", labels(map[string]string{"one": "value", "two": "value"})),
	}

	assert.Equal(t, want, (&core{}).withLabels(fields))
}

func TestExtractLabels(t *testing.T) {
	var lbls labels
	c := &core{zapcore.NewNopCore(), labels{}, labels{}}

	fields := []zap.Field{
		zap.String("hello", "world"),
		Label("one", "world"),
		Label("two", "worlds"),
	}

	lbls, fields = c.extractLabels(fields)

	require.Len(t, lbls, 2)
	assert.Equal(t, "world", lbls["one"])
	assert.Equal(t, "worlds", lbls["two"])

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
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := &core{debugcore, labels{}, labels{}}

	fields := []zap.Field{
		zap.String("hello", "world"),
		Label("one", "value"),
		Label("two", "value"),
	}

	err := core.Write(zapcore.Entry{}, fields)
	require.NoError(t, err)

	assert.NotNil(t, logs.All()[0].ContextMap()["labels"])
}

func TestWithAndWrite(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{debugcore, labels{}, labels{}})

	core = core.With([]zapcore.Field{Label("one", "world")})
	err := core.Write(zapcore.Entry{}, []zapcore.Field{Label("two", "worlds")})
	require.NoError(t, err)

	labels := logs.All()[0].ContextMap()["labels"].(map[string]interface{})

	assert.Equal(t, "world", labels["one"])
	assert.Equal(t, "worlds", labels["two"])
}

func TestWithAndWrite_MultipleEntries(t *testing.T) {
	debugcore, logs := observer.New(zapcore.DebugLevel)
	core := zapcore.Core(&core{debugcore, labels{}, labels{}})

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
