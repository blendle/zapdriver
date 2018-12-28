package zapdriver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceContext(t *testing.T) {
	t.Parallel()

	got := ServiceContext("test service name").Interface.(*serviceContext)

	assert.Equal(t, "test service name", got.Name)
}

func TestNewServiceContext(t *testing.T) {
	t.Parallel()

	got := newServiceContext("test service name")

	assert.Equal(t, "test service name", got.Name)
}
