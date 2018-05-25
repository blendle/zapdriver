package zapdriver

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	assert.Equal(t, want, (&Core{}).withLabels(fields))
}

func TestWithSourceLocation(t *testing.T) {
	fields := []zap.Field{zap.String("hello", "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}

	want := []zap.Field{
		zap.String("hello", "world"),
		zap.Object("sourceLocation", newSource(pc, file, line, ok)),
	}

	assert.Equal(t, want, (&Core{}).withSourceLocation(ent, fields))
}

func TestWithSourceLocation_DoesNotOverwrite(t *testing.T) {
	fields := []zap.Field{zap.String("sourceLocation", "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}

	want := []zap.Field{
		zap.String("sourceLocation", "world"),
	}

	assert.Equal(t, want, (&Core{}).withSourceLocation(ent, fields))
}

func TestWithSourceLocation_OnlyWhenDefined(t *testing.T) {
	fields := []zap.Field{zap.String("hello", "world")}
	pc, file, line, ok := runtime.Caller(0)
	ent := zapcore.Entry{Caller: zapcore.NewEntryCaller(pc, file, line, ok)}
	ent.Caller.Defined = false

	want := []zap.Field{
		zap.String("hello", "world"),
	}

	assert.Equal(t, want, (&Core{}).withSourceLocation(ent, fields))
}
