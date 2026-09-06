package supabase

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestNullableUUID(t *testing.T) {
	t.Run("zero maps to nil", func(t *testing.T) {
		require.Nil(t, nullableUUID(uuid.UUID{}))
	})

	t.Run("non-zero passes through", func(t *testing.T) {
		id := uuid.NewV4()
		require.Equal(t, id, nullableUUID(id))
	})
}
