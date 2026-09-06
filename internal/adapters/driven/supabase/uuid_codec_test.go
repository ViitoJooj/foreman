package supabase

import (
	"testing"
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func newRegisteredMap() *pgtype.Map {
	m := pgtype.NewMap()
	registerUUIDType(m)
	return m
}

func roundTrip(t *testing.T, format int16, in uuid.UUID) uuid.UUID {
	t.Helper()
	m := newRegisteredMap()

	encoded, err := m.Encode(pgtype.UUIDOID, format, in, nil)
	require.NoError(t, err)

	var out uuid.UUID
	require.NoError(t, m.Scan(pgtype.UUIDOID, format, encoded, &out))
	return out
}

func TestUUIDCodecBinaryRoundTrip(t *testing.T) {
	in := uuid.NewV4()
	require.Equal(t, in, roundTrip(t, pgtype.BinaryFormatCode, in))
}

func TestUUIDCodecTextRoundTrip(t *testing.T) {
	in := uuid.NewV4()
	require.Equal(t, in, roundTrip(t, pgtype.TextFormatCode, in))
}

func TestUUIDCodecZeroValueRoundTrip(t *testing.T) {
	require.Equal(t, uuid.UUID{}, roundTrip(t, pgtype.BinaryFormatCode, uuid.UUID{}))
}

func TestUUIDCodecScansNullAsZeroValue(t *testing.T) {
	m := newRegisteredMap()

	var out uuid.UUID
	require.NoError(t, m.Scan(pgtype.UUIDOID, pgtype.BinaryFormatCode, nil, &out))
	require.Equal(t, uuid.UUID{}, out)
}
