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

	labels := newLabels()
	labels.store = map[string]string{"hello": "world", "hi": "universe"}

	assert.Equal(t, zap.Object(labelsKey, labels), field)
}
