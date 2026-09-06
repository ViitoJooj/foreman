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
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestAgentHandlerCreate(t *testing.T) {
	a := newTestAPI(t)
	companyID := testutil.RandomUUID()
	name := testutil.RandomAgentName()

	a.agentRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ag domain.Agent) (domain.Agent, error) {
			require.Equal(t, companyID, ag.CompanyID)
			require.Equal(t, name, ag.Name)
			require.Equal(t, domain.RoleCoder, ag.Role)
			require.Equal(t, domain.AgentIdle, ag.Status)
			ag.ID = testutil.RandomUUID()
			ag.CreatedAt = time.Now()
			ag.UpdatedAt = ag.CreatedAt
			return ag, nil
		})

	rec := a.do(t, http.MethodPost, "/api/v1/agents", map[string]any{
		"company_id": companyID.String(),
		"name":       name,
		"role":       string(domain.RoleCoder),
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "idle", resp["status"])
}

func TestAgentHandlerCreateRejectsUnknownRole(t *testing.T) {
	a := newTestAPI(t)

	rec := a.do(t, http.MethodPost, "/api/v1/agents", map[string]any{
		"company_id": testutil.RandomUUID().String(),
		"name":       testutil.RandomAgentName(),
		"role":       testutil.RandomSlug(),
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAgentHandlerGet(t *testing.T) {
	a := newTestAPI(t)
	id := testutil.RandomUUID()

	a.agentRepo.EXPECT().Get(gomock.Any(), id).Return(domain.Agent{
		ID:     id,
		Name:   testutil.RandomAgentName(),
		Role:   domain.RoleTester,
		Status: domain.AgentWorking,
	}, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/agents/"+id.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "tester", resp["role"])
}

func TestAgentHandlerListRequiresCompany(t *testing.T) {
	a := newTestAPI(t)

	rec := a.do(t, http.MethodGet, "/api/v1/agents", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAgentHandlerListByCompany(t *testing.T) {
	a := newTestAPI(t)
	companyID := testutil.RandomUUID()

	a.agentRepo.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Agent{
		{ID: testutil.RandomUUID(), CompanyID: companyID, Name: testutil.RandomAgentName(), Role: domain.RoleCoder, Status: domain.AgentIdle},
	}, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/agents?company="+companyID.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
}
