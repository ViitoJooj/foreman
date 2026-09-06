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

func TestCompanyHandlerCreate(t *testing.T) {
	a := newTestAPI(t)
	name := testutil.RandomCompanyName()
	slug := testutil.RandomSlug()

	a.companyRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.Company) (domain.Company, error) {
			require.Equal(t, name, c.Name)
			require.Equal(t, slug, c.Slug)
			c.ID = testutil.RandomUUID()
			c.CreatedAt = time.Now()
			c.UpdatedAt = c.CreatedAt
			return c, nil
		})

	rec := a.do(t, http.MethodPost, "/api/v1/companies", map[string]any{
		"name":       name,
		"slug":       slug,
		"repo_owner": testutil.RandomRepoOwner(),
		"repo_name":  testutil.RandomRepoName(),
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, name, resp["name"])
	require.NotEmpty(t, resp["id"])
}

func TestCompanyHandlerCreateRejectsMissingFields(t *testing.T) {
	a := newTestAPI(t)

	rec := a.do(t, http.MethodPost, "/api/v1/companies", map[string]any{
		"name": testutil.RandomCompanyName(),
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCompanyHandlerGet(t *testing.T) {
	a := newTestAPI(t)
	id := testutil.RandomUUID()

	a.companyRepo.EXPECT().Get(gomock.Any(), id).Return(domain.Company{
		ID:   id,
		Name: testutil.RandomCompanyName(),
		Slug: testutil.RandomSlug(),
	}, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/companies/"+id.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, id.String(), resp["id"])
}

func TestCompanyHandlerGetNotFound(t *testing.T) {
	a := newTestAPI(t)
	id := testutil.RandomUUID()

	a.companyRepo.EXPECT().Get(gomock.Any(), id).Return(domain.Company{}, ports.ErrCompanyNotFound)

	rec := a.do(t, http.MethodGet, "/api/v1/companies/"+id.String(), nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCompanyHandlerList(t *testing.T) {
	a := newTestAPI(t)

	a.companyRepo.EXPECT().List(gomock.Any()).Return([]domain.Company{
		{ID: testutil.RandomUUID(), Name: testutil.RandomCompanyName(), Slug: testutil.RandomSlug()},
		{ID: testutil.RandomUUID(), Name: testutil.RandomCompanyName(), Slug: testutil.RandomSlug()},
	}, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/companies", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp, 2)
}
