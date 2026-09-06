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

func TestTaskHandlerCreate(t *testing.T) {
	a := newTestAPI(t)
	companyID := testutil.RandomUUID()
	title := testutil.RandomTaskTitle()

	a.taskRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, task domain.Task) (domain.Task, error) {
			require.Equal(t, companyID, task.CompanyID)
			require.Equal(t, title, task.Title)
			require.Equal(t, domain.TaskQueued, task.State)
			task.ID = testutil.RandomUUID()
			task.CreatedAt = time.Now()
			task.UpdatedAt = task.CreatedAt
			return task, nil
		})

	rec := a.do(t, http.MethodPost, "/api/v1/tasks", map[string]any{
		"company_id": companyID.String(),
		"title":      title,
		"risk":       string(domain.RiskLow),
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, companyID.String(), resp["company_id"])
	require.Equal(t, "queued", resp["state"])
}

func TestTaskHandlerCreateRejectsMalformedBody(t *testing.T) {
	a := newTestAPI(t)
	// No repository call expected: binding fails first.

	rec := a.do(t, http.MethodPost, "/api/v1/tasks", map[string]any{
		"company_id": testutil.RandomSlug(),
		"title":      testutil.RandomTaskTitle(),
		"risk":       "low",
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskHandlerListRequiresFilter(t *testing.T) {
	a := newTestAPI(t)

	rec := a.do(t, http.MethodGet, "/api/v1/tasks", nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskHandlerListByCompany(t *testing.T) {
	a := newTestAPI(t)
	companyID := testutil.RandomUUID()

	a.taskRepo.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Task{
		{
			ID:        testutil.RandomUUID(),
			CompanyID: companyID,
			Title:     testutil.RandomTaskTitle(),
			State:     domain.TaskCoding,
			Risk:      domain.RiskLow,
		},
	}, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/tasks?company="+companyID.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
	require.Equal(t, companyID.String(), resp[0]["company_id"])
}
