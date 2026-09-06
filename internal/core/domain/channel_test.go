package domain_test

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestChannelHoldsAssignedFields(t *testing.T) {
	id := testutil.RandomUUID()
	companyID := testutil.RandomUUID()
	name := testutil.RandomChannelName()
	created := time.Now()

	ch := domain.Channel{ID: id, CompanyID: companyID, Name: name, CreatedAt: created}

	require.Equal(t, id, ch.ID)
	require.Equal(t, companyID, ch.CompanyID)
	require.Equal(t, name, ch.Name)
	require.Equal(t, created, ch.CreatedAt)
}

func TestChannelZeroValue(t *testing.T) {
	var ch domain.Channel
	require.Equal(t, uuid.UUID{}, ch.ID)
	require.Empty(t, ch.Name)
	require.True(t, ch.CreatedAt.IsZero())
}
