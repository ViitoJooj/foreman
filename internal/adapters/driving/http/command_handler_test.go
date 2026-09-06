package http

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestCommandHandlerCreateKill(t *testing.T) {
	a := newTestAPI(t)
	companyID := testutil.RandomUUID()

	a.commandRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.Command) (domain.Command, error) {
			require.Equal(t, domain.CommandKill, c.Kind)
			require.Equal(t, companyID, c.CompanyID)
			require.Equal(t, domain.CommandPending, c.Status)
			c.ID = testutil.RandomUUID()
			c.CreatedAt = time.Now()
			return c, nil
		})

	rec := a.do(t, http.MethodPost, "/api/v1/commands", map[string]any{
		"kind":       "kill",
		"company_id": companyID.String(),
		"reason":     testutil.RandomText(),
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "kill", resp["kind"])
	require.Equal(t, "pending", resp["status"])
}

func TestCommandHandlerCreatePause(t *testing.T) {
	a := newTestAPI(t)
	agentID := testutil.RandomUUID()

	a.commandRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.Command) (domain.Command, error) {
			require.Equal(t, domain.CommandPause, c.Kind)
			require.Equal(t, agentID, c.TargetAgentID)
			c.ID = testutil.RandomUUID()
			return c, nil
		})

	rec := a.do(t, http.MethodPost, "/api/v1/commands", map[string]any{
		"kind":            "pause",
		"target_agent_id": agentID.String(),
	})

	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestCommandHandlerCreateRejectsInconsistentTarget(t *testing.T) {
	a := newTestAPI(t)
	// kill must not carry a target agent -> service rejects, no repo call.

	rec := a.do(t, http.MethodPost, "/api/v1/commands", map[string]any{
		"kind":            "kill",
		"target_agent_id": testutil.RandomUUID().String(),
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommandHandlerCreateRejectsUnknownKind(t *testing.T) {
	a := newTestAPI(t)

	rec := a.do(t, http.MethodPost, "/api/v1/commands", map[string]any{
		"kind": testutil.RandomSlug(),
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommandHandlerListDefaultsToPending(t *testing.T) {
	a := newTestAPI(t)

	a.commandRepo.EXPECT().ListByStatus(gomock.Any(), domain.CommandPending).Return([]domain.Command{
		{ID: testutil.RandomUUID(), Kind: domain.CommandKill, Status: domain.CommandPending},
	}, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/commands", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
}

func TestCommandHandlerListHonoursStatusQuery(t *testing.T) {
	a := newTestAPI(t)

	a.commandRepo.EXPECT().ListByStatus(gomock.Any(), domain.CommandApplied).Return(nil, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/commands?status=applied", nil)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestCommandHandlerGetNotFound(t *testing.T) {
	a := newTestAPI(t)
	id := testutil.RandomUUID()

	a.commandRepo.EXPECT().Get(gomock.Any(), id).Return(domain.Command{}, ports.ErrCommandNotFound)

	rec := a.do(t, http.MethodGet, "/api/v1/commands/"+id.String(), nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
