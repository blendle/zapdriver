package zapdriver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

func TestNewProduction(t *testing.T) {
	logger, err := NewProduction(zap.Fields(zap.String("hello", "world")))

	require.NoError(t, err)
	assert.IsType(t, &zap.Logger{}, logger)
}

func TestNewProductionWithCore(t *testing.T) {
	logger, err := NewProductionWithCore(
		WrapCore(ReportAllErrors(true)),
		zap.Fields(zap.String("hello", "world")),
	)

	require.NoError(t, err)
	assert.IsType(t, &zap.Logger{}, logger)
}

func TestNewDevelopment(t *testing.T) {
	logger, err := NewDevelopment(zap.Fields(zap.String("hello", "world")))

	require.NoError(t, err)
	assert.IsType(t, &zap.Logger{}, logger)
}

func TestNewDevelopmentWithCore(t *testing.T) {
	logger, err := NewDevelopmentWithCore(
		WrapCore(ReportAllErrors(true)),
		zap.Fields(zap.String("hello", "world")),
	)

	require.NoError(t, err)
	assert.IsType(t, &zap.Logger{}, logger)
}
