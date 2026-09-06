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

func TestChannelHandlerCreate(t *testing.T) {
	a := newTestAPI(t)
	companyID := testutil.RandomUUID()
	name := testutil.RandomChannelName()

	a.channelRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ch domain.Channel) (domain.Channel, error) {
			require.Equal(t, companyID, ch.CompanyID)
			require.Equal(t, name, ch.Name)
			ch.ID = testutil.RandomUUID()
			ch.CreatedAt = time.Now()
			return ch, nil
		})

	rec := a.do(t, http.MethodPost, "/api/v1/channels", map[string]any{
		"company_id": companyID.String(),
		"name":       name,
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, name, resp["name"])
	require.Equal(t, companyID.String(), resp["company_id"])
}

func TestChannelHandlerCreateRejectsInvalidCompanyID(t *testing.T) {
	a := newTestAPI(t)

	rec := a.do(t, http.MethodPost, "/api/v1/channels", map[string]any{
		"company_id": testutil.RandomSlug(),
		"name":       testutil.RandomChannelName(),
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestChannelHandlerGet(t *testing.T) {
	a := newTestAPI(t)
	id := testutil.RandomUUID()

	a.channelRepo.EXPECT().Get(gomock.Any(), id).Return(domain.Channel{
		ID:   id,
		Name: testutil.RandomChannelName(),
	}, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/channels/"+id.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, id.String(), resp["id"])
}

func TestChannelHandlerListByCompany(t *testing.T) {
	a := newTestAPI(t)
	companyID := testutil.RandomUUID()

	a.channelRepo.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Channel{
		{ID: testutil.RandomUUID(), CompanyID: companyID, Name: testutil.RandomChannelName()},
	}, nil)

	rec := a.do(t, http.MethodGet, "/api/v1/channels?company="+companyID.String(), nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
}

// TestChannelRouteDoesNotShadowMessages guards the shared /channels/:id prefix:
// GET /channels/:id and POST /channels/:id/messages must coexist.
func TestChannelRouteDoesNotShadowMessages(t *testing.T) {
	a := newTestAPI(t)
	channelID := testutil.RandomUUID()
	fromID := testutil.RandomUUID()

	a.msgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, m domain.Message) (domain.Message, error) {
			m.ID = testutil.RandomUUID()
			m.CreatedAt = time.Now()
			return m, nil
		})
	a.msgBus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

	rec := a.do(t, http.MethodPost, "/api/v1/channels/"+channelID.String()+"/messages", map[string]any{
		"type":          string(domain.MessageStatus),
		"from_agent_id": fromID.String(),
	})

	require.Equal(t, http.StatusCreated, rec.Code)
}
