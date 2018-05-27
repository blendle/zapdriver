package zapdriver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestOperation(t *testing.T) {
	t.Parallel()

	op := &operation{ID: "id", Producer: "producer", First: true, Last: false}
	field := Operation("id", "producer", true, false)

	assert.Equal(t, zap.Object(operationKey, op), field)
}
