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

func TestOperationStart(t *testing.T) {
	t.Parallel()

	op := &operation{ID: "id", Producer: "producer", First: true, Last: false}
	field := OperationStart("id", "producer")

	assert.Equal(t, zap.Object(operationKey, op), field)
}

func TestOperationCont(t *testing.T) {
	t.Parallel()

	op := &operation{ID: "id", Producer: "producer", First: false, Last: false}
	field := OperationCont("id", "producer")

	assert.Equal(t, zap.Object(operationKey, op), field)
}

func TestOperationEnd(t *testing.T) {
	t.Parallel()

	op := &operation{ID: "id", Producer: "producer", First: false, Last: true}
	field := OperationEnd("id", "producer")

	assert.Equal(t, zap.Object(operationKey, op), field)
}
