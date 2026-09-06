package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/service"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func init() { gin.SetMode(gin.TestMode) }

// testAPI wires the real router and use cases against mocked ports so tests can
// drive full HTTP requests while controlling the persistence layer.
type testAPI struct {
	router      *gin.Engine
	apiKey      string
	companyRepo *testutil.MockCompanyRepository
	agentRepo   *testutil.MockAgentRepository
	channelRepo *testutil.MockChannelRepository
	taskRepo    *testutil.MockTaskRepository
	msgRepo     *testutil.MockMessageRepository
	msgBus      *testutil.MockMessageBus
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	ctrl := gomock.NewController(t)

	a := &testAPI{
		apiKey:      testutil.RandomAPIKey(),
		companyRepo: testutil.NewMockCompanyRepository(ctrl),
		agentRepo:   testutil.NewMockAgentRepository(ctrl),
		channelRepo: testutil.NewMockChannelRepository(ctrl),
		taskRepo:    testutil.NewMockTaskRepository(ctrl),
		msgRepo:     testutil.NewMockMessageRepository(ctrl),
		msgBus:      testutil.NewMockMessageBus(ctrl),
	}
	a.router = NewRouter(a.apiKey, Handlers{
		Companies: NewCompanyHandler(service.NewCompany(a.companyRepo)),
		Agents:    NewAgentHandler(service.NewAgent(a.agentRepo)),
		Channels:  NewChannelHandler(service.NewChannel(a.channelRepo)),
		Tasks:     NewTaskHandler(service.NewCreateTask(a.taskRepo), service.NewListTasks(a.taskRepo)),
		Messages:  NewMessageHandler(service.NewPostMessage(a.msgRepo, a.msgBus)),
	})
	return a
}

func (a *testAPI) do(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("X-Api-Key", a.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	a.router.ServeHTTP(rec, req)
	return rec
}

func TestNewRouterUnknownRouteReturns404(t *testing.T) {
	a := newTestAPI(t)
	rec := a.do(t, http.MethodGet, "/does-not-exist", nil)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNewRouterGuardsV1WithAPIKey(t *testing.T) {
	a := newTestAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil) // no X-Api-Key
	rec := httptest.NewRecorder()
	a.router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
