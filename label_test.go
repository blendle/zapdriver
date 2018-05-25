package zapdriver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLabel(t *testing.T) {
	t.Parallel()

	field := Label("key", "value")

	assert.Equal(t, zap.String("labels.key", "value"), field)
}

func TestLabels(t *testing.T) {
	t.Parallel()

	field := Labels(
		Label("hello", "world"),
		Label("hi", "universe"),
	)

	assert.Equal(t, zap.Object("labels", labels{"hello": "world", "hi": "universe"}), field)
}
