package zapdriver

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Core is a zapdriver specific core wrapped around the default zap core. It
// allows to merge all defined labels
type Core struct {
	zapcore.Core
}

func withCore() zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &Core{core}
	})
}

// With adds structured context to the Core.
func (c *Core) With(fields []zap.Field) zapcore.Core {
	return &Core{c.Core.With(fields)}
}

// Check determines whether the supplied Entry should be logged (using the
// embedded LevelEnabler and possibly some extra logic). If the entry
// should be logged, the Core adds itself to the CheckedEntry and returns
// the result.
//
// Callers must use Check before calling Write.
func (c *Core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

func (c *Core) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	fields = c.withLabels(fields)
	fields = c.withSourceLocation(ent, fields)

	return c.Core.Write(ent, fields)
}

// Sync flushes buffered logs (if any).
func (c *Core) Sync() error {
	return c.Core.Sync()
}

func (c *Core) withLabels(fields []zapcore.Field) []zapcore.Field {
	labels := labels{}
	out := []zapcore.Field{}

	for i := range fields {
		if isLabelField(fields[i]) {
			labels[strings.Replace(fields[i].Key, "labels.", "", 1)] = fields[i].String
			continue
		}

		out = append(out, fields[i])
	}

	return append(out, labelsField(labels))
}

func (c *Core) withSourceLocation(ent zapcore.Entry, fields []zapcore.Field) []zapcore.Field {
	// If the source location was manually set, don't overwrite it
	for i := range fields {
		if fields[i].Key == "sourceLocation" {
			return fields
		}
	}

	if !ent.Caller.Defined {
		return fields
	}

	return append(fields, SourceLocation(ent.Caller.PC, ent.Caller.File, ent.Caller.Line, true))
}
