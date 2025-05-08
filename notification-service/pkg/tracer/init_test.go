package tracer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitTracer_Success(t *testing.T) {
	shutdown, err := InitTracer("test_service")
	require.NoError(t, err)
	require.NotNil(t, shutdown)

	err = shutdown(context.Background())
	require.NoError(t, err)
}
