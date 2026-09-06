//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/bus"
	"github.com/ViitoJooj/foreman/internal/adapters/driven/supabase"
	httpapi "github.com/ViitoJooj/foreman/internal/adapters/driving/http"
	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
	"github.com/ViitoJooj/foreman/internal/orchestrator"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func drainPendingCommands(t *testing.T, ctx context.Context, repo *supabase.CommandRepository) {
	t.Helper()
	pending, err := repo.ListByStatus(ctx, domain.CommandPending)
	require.NoError(t, err)
	for _, cmd := range pending {
		cmd.Status = domain.CommandApplied
		now := time.Now().UTC()
		cmd.AppliedAt = &now
		_, err := repo.Update(ctx, cmd)
		require.NoError(t, err)
	}
}

type apiCaller struct {
	t      *testing.T
	base   string
	apiKey string
}

func (c apiCaller) do(method, path string, body any) (int, []byte) {
	c.t.Helper()
	var r io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(c.t, err)
		r = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.base+path, r)
	require.NoError(c.t, err)
	req.Header.Set("X-Api-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(c.t, err)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

func idOf(t *testing.T, data []byte) string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	id, _ := m["id"].(string)
	require.NotEmpty(t, id)
	return id
}

func TestAPIEndToEnd(t *testing.T) {
	companies := supabase.NewCompanyRepository(testPool)
	agents := supabase.NewAgentRepository(testPool)
	channels := supabase.NewChannelRepository(testPool)
	tasks := supabase.NewTaskRepository(testPool)
	messages := supabase.NewMessageRepository(testPool)
	commands := supabase.NewCommandRepository(testPool)
	messageBus := bus.NewInProcess()

	gin.SetMode(gin.TestMode)
	apiKey := testutil.RandomAPIKey()
	router := httpapi.NewRouter(apiKey, httpapi.Handlers{
		Companies: httpapi.NewCompanyHandler(service.NewCompany(companies)),
		Agents:    httpapi.NewAgentHandler(service.NewAgent(agents)),
		Channels:  httpapi.NewChannelHandler(service.NewChannel(channels)),
		Tasks:     httpapi.NewTaskHandler(service.NewCreateTask(tasks), service.NewListTasks(tasks)),
		Messages:  httpapi.NewMessageHandler(service.NewPostMessage(messages, messageBus), service.NewTailMessages(messages, messageBus)),
		Commands:  httpapi.NewCommandHandler(service.NewCommand(commands)),
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	api := apiCaller{t: t, base: srv.URL, apiKey: apiKey}

	// Auth guard.
	code, _ := apiCaller{t: t, base: srv.URL, apiKey: "wrong"}.do(http.MethodGet, "/api/v1/companies", nil)
	require.Equal(t, http.StatusUnauthorized, code)

	// Company.
	code, body := api.do(http.MethodPost, "/api/v1/companies", map[string]any{
		"name": testutil.RandomCompanyName(), "slug": testutil.RandomSlug(),
		"repo_owner": testutil.RandomRepoOwner(), "repo_name": testutil.RandomRepoName(),
	})
	require.Equal(t, http.StatusCreated, code, string(body))
	companyID := idOf(t, body)

	code, body = api.do(http.MethodGet, "/api/v1/companies/"+companyID, nil)
	require.Equal(t, http.StatusOK, code, string(body))

	// Agent.
	code, body = api.do(http.MethodPost, "/api/v1/agents", map[string]any{
		"company_id": companyID, "name": testutil.RandomAgentName(), "role": "coder",
	})
	require.Equal(t, http.StatusCreated, code, string(body))
	agentID := idOf(t, body)

	code, body = api.do(http.MethodGet, "/api/v1/agents?company="+companyID, nil)
	require.Equal(t, http.StatusOK, code)
	var agentList []map[string]any
	require.NoError(t, json.Unmarshal(body, &agentList))
	require.Len(t, agentList, 1)

	// Channel.
	code, body = api.do(http.MethodPost, "/api/v1/channels", map[string]any{
		"company_id": companyID, "name": testutil.RandomChannelName(),
	})
	require.Equal(t, http.StatusCreated, code, string(body))
	channelID := idOf(t, body)

	// Task enters queued.
	code, body = api.do(http.MethodPost, "/api/v1/tasks", map[string]any{
		"company_id": companyID, "title": testutil.RandomTaskTitle(), "risk": "low",
	})
	require.Equal(t, http.StatusCreated, code, string(body))
	var taskResp map[string]any
	require.NoError(t, json.Unmarshal(body, &taskResp))
	require.Equal(t, "queued", taskResp["state"])

	// Message.
	code, body = api.do(http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{
		"type": "status", "from_agent_id": agentID, "body": testutil.RandomText(),
		"payload": map[string]any{"k": 1},
	})
	require.Equal(t, http.StatusCreated, code, string(body))

	// Command. Use pause (agent-scoped) rather than kill so it does not engage
	// the global stop for other tests sharing this database.
	code, body = api.do(http.MethodPost, "/api/v1/commands", map[string]any{
		"kind": "pause", "target_agent_id": agentID,
	})
	require.Equal(t, http.StatusCreated, code, string(body))
	code, body = api.do(http.MethodGet, "/api/v1/commands?status=pending", nil)
	require.Equal(t, http.StatusOK, code)
	var pending []map[string]any
	require.NoError(t, json.Unmarshal(body, &pending))
	require.GreaterOrEqual(t, len(pending), 1)
}

func TestOrchestratorEndToEnd(t *testing.T) {
	ctx := context.Background()
	companyRepo := supabase.NewCompanyRepository(testPool)
	agentRepo := supabase.NewAgentRepository(testPool)
	channelRepo := supabase.NewChannelRepository(testPool)
	taskRepo := supabase.NewTaskRepository(testPool)
	messageRepo := supabase.NewMessageRepository(testPool)
	commandRepo := supabase.NewCommandRepository(testPool)
	messageBus := bus.NewInProcess()

	// Drain any pending commands other tests left behind so this cycle is not
	// affected by a stray global kill.
	drainPendingCommands(t, ctx, commandRepo)

	company, err := companyRepo.Create(ctx, domain.Company{
		Name: testutil.RandomCompanyName(), Slug: testutil.RandomSlug(),
		RepoOwner: testutil.RandomRepoOwner(), RepoName: testutil.RandomRepoName(),
	})
	require.NoError(t, err)

	for _, role := range []domain.AgentRole{domain.RoleCoder, domain.RoleTester, domain.RolePRReviewer} {
		_, err := agentRepo.Create(ctx, domain.Agent{
			CompanyID: company.ID, Name: testutil.RandomAgentName(), Role: role, Status: domain.AgentIdle,
		})
		require.NoError(t, err)
	}
	_, err = channelRepo.Create(ctx, domain.Channel{CompanyID: company.ID, Name: testutil.RandomChannelName()})
	require.NoError(t, err)

	task, err := taskRepo.Create(ctx, domain.Task{
		CompanyID: company.ID, Title: testutil.RandomTaskTitle(),
		State: domain.TaskQueued, Risk: domain.RiskLow, MaxRetries: 3,
	})
	require.NoError(t, err)

	orch := orchestrator.New(orchestrator.Deps{
		Tasks: taskRepo, Agents: agentRepo, Channels: channelRepo,
		Messages: messageRepo, Bus: messageBus, Commands: commandRepo,
		Runners: orchestrator.StubRunners(), Interval: time.Hour, Logger: quiet(),
	})

	// Stub runners cascade a low-risk task through the whole pipeline in one cycle.
	require.NoError(t, orch.Cycle(ctx))

	advanced, err := taskRepo.Get(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskMerged, advanced.State)
	require.NotEqual(t, task.AssigneeID, advanced.AssigneeID)
}
