package zapdriver_test

import (
	"testing"

	"github.com/blendle/zapdriver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestEncodeLevel(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		lvl  zapcore.Level
		want string
	}{
		{zapcore.DebugLevel, "DEBUG"},
		{zapcore.InfoLevel, "INFO"},
		{zapcore.WarnLevel, "WARNING"},
		{zapcore.ErrorLevel, "ERROR"},
		{zapcore.DPanicLevel, "CRITICAL"},
		{zapcore.PanicLevel, "ALERT"},
		{zapcore.FatalLevel, "EMERGENCY"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			enc := &sliceArrayEncoder{}
			zapdriver.EncodeLevel(tt.lvl, enc)

			require.Len(t, enc.elems, 1)
			assert.Equal(t, enc.elems[0].(string), tt.want)
		})
	}
}
