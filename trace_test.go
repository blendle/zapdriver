package zapdriver

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestTraceContext(t *testing.T) {
	t.Parallel()

	fields := TraceContext("105445aa7843bc8bf206b120001000", "0", true, "my-project-name")
	assert.Equal(t, fields, []zap.Field{
		zap.String(traceKey, "projects/my-project-name/traces/105445aa7843bc8bf206b120001000"),
		zap.String(spanKey, "0"),
		zap.String(traceSampledKey, "true"),
	})
}
