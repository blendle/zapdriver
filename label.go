package zapdriver

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Label adds an optional label to the payload.
//
// Labels are a set of user-defined (key, value) data that provides additional
// information about the log entry.
//
// Example: { "name": "wrench", "mass": "1.3kg", "count": "3" }.
func Label(key, value string) zap.Field {
	return zap.String("labels."+key, value)
}

// Labels takes Zap fields, filters the ones that have their key start with the
// string `labels.` and their value type set to StringType. It then wraps those
// key/value pairs in a top-level `labels` namespace.
func Labels(fields ...zap.Field) zap.Field {
	lbls := labels{}

	for i := range fields {
		if isLabelField(fields[i]) {
			lbls[strings.Replace(fields[i].Key, "labels.", "", 1)] = fields[i].String
		}
	}

	return labelsField(lbls)
}

func isLabelField(field zap.Field) bool {
	return strings.HasPrefix(field.Key, "labels.") && field.Type == zapcore.StringType
}

func labelsField(l map[string]string) zap.Field {
	return zap.Object("labels", labels(l))
}

type labels map[string]string

func (l labels) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range l {
		enc.AddString(k, v)
	}

	return nil
}
