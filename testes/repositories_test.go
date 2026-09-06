//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/supabase"
	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func seedCompany(t *testing.T) domain.Company {
	t.Helper()
	c, err := supabase.NewCompanyRepository(testPool).Create(context.Background(), domain.Company{
		Name:      testutil.RandomCompanyName(),
		Slug:      testutil.RandomSlug(),
		RepoOwner: testutil.RandomRepoOwner(),
		RepoName:  testutil.RandomRepoName(),
	})
	require.NoError(t, err)
	return c
}

func seedAgent(t *testing.T, companyID uuid.UUID) domain.Agent {
	t.Helper()
	a, err := supabase.NewAgentRepository(testPool).Create(context.Background(), domain.Agent{
		CompanyID: companyID,
		Name:      testutil.RandomAgentName(),
		Role:      testutil.RandomAgentRole(),
		Status:    domain.AgentIdle,
	})
	require.NoError(t, err)
	return a
}

func seedChannel(t *testing.T, companyID uuid.UUID) domain.Channel {
	t.Helper()
	ch, err := supabase.NewChannelRepository(testPool).Create(context.Background(), domain.Channel{
		CompanyID: companyID,
		Name:      testutil.RandomChannelName(),
	})
	require.NoError(t, err)
	return ch
}

func TestCompanyRepository(t *testing.T) {
	ctx := context.Background()
	repo := supabase.NewCompanyRepository(testPool)

	created := seedCompany(t)
	require.NotEqual(t, uuid.UUID{}, created.ID)
	require.False(t, created.CreatedAt.IsZero())

	got, err := repo.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created, got)

	created.Name = testutil.RandomCompanyName()
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, created.Name, updated.Name)
	require.False(t, updated.UpdatedAt.Before(got.UpdatedAt), "updated_at trigger moved the timestamp forward")

	all, err := repo.List(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, all)

	_, err = repo.Get(ctx, testutil.RandomUUID())
	require.ErrorIs(t, err, ports.ErrCompanyNotFound)
}

func TestAgentRepository(t *testing.T) {
	ctx := context.Background()
	repo := supabase.NewAgentRepository(testPool)
	company := seedCompany(t)

	created := seedAgent(t, company.ID)

	got, err := repo.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created, got)

	created.Status = domain.AgentWorking
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, domain.AgentWorking, updated.Status)

	list, err := repo.ListByCompany(ctx, company.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	_, err = repo.Get(ctx, testutil.RandomUUID())
	require.ErrorIs(t, err, ports.ErrAgentNotFound)
}

func TestChannelRepository(t *testing.T) {
	ctx := context.Background()
	repo := supabase.NewChannelRepository(testPool)
	company := seedCompany(t)

	created := seedChannel(t, company.ID)

	got, err := repo.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created, got)

	list, err := repo.ListByCompany(ctx, company.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	_, err = repo.Get(ctx, testutil.RandomUUID())
	require.ErrorIs(t, err, ports.ErrChannelNotFound)
}

func TestTaskRepository(t *testing.T) {
	ctx := context.Background()
	repo := supabase.NewTaskRepository(testPool)
	company := seedCompany(t)
	agent := seedAgent(t, company.ID)

	created, err := repo.Create(ctx, domain.Task{
		CompanyID:   company.ID,
		Title:       testutil.RandomTaskTitle(),
		Description: testutil.RandomText(),
		State:       domain.TaskCreated,
		Risk:        domain.RiskLow,
		MaxRetries:  3,
	})
	require.NoError(t, err)
	require.Equal(t, domain.TaskCreated, created.State)
	require.Equal(t, uuid.UUID{}, created.AssigneeID, "unset assignee round-trips as the zero UUID (SQL NULL)")

	got, err := repo.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created, got)

	created.State = domain.TaskCoding
	created.Branch = testutil.RandomBranchSlug()
	created.AssigneeID = agent.ID
	created.Retries = 1
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, domain.TaskCoding, updated.State)
	require.Equal(t, agent.ID, updated.AssigneeID)
	require.Equal(t, created.Branch, updated.Branch)

	byCompany, err := repo.ListByCompany(ctx, company.ID)
	require.NoError(t, err)
	require.Len(t, byCompany, 1)

	byState, err := repo.ListByState(ctx, domain.TaskCoding)
	require.NoError(t, err)
	require.NotEmpty(t, byState)

	_, err = repo.Get(ctx, testutil.RandomUUID())
	require.ErrorIs(t, err, ports.ErrTaskNotFound)
}

func TestMessageRepository(t *testing.T) {
	ctx := context.Background()
	repo := supabase.NewMessageRepository(testPool)
	company := seedCompany(t)
	channel := seedChannel(t, company.ID)
	sender := seedAgent(t, company.ID)
	recipient := seedAgent(t, company.ID)

	payload := testutil.RandomJSONPayload()
	first, err := repo.Create(ctx, domain.Message{
		ChannelID:        channel.ID,
		Type:             domain.MessageRequest,
		FromAgentID:      sender.ID,
		ToAgentID:        recipient.ID,
		RequiresResponse: true,
		Body:             testutil.RandomText(),
		Payload:          payload,
	})
	require.NoError(t, err)
	require.JSONEq(t, string(payload), string(first.Payload), "jsonb payload round-trips verbatim")

	got, err := repo.Get(ctx, first.ID)
	require.NoError(t, err)
	require.Equal(t, first, got)

	// A status message with no recipient and no reply target.
	second, err := repo.Create(ctx, domain.Message{
		ChannelID:   channel.ID,
		Type:        domain.MessageStatus,
		FromAgentID: sender.ID,
		Body:        testutil.RandomText(),
	})
	require.NoError(t, err)
	require.Equal(t, uuid.UUID{}, second.ToAgentID)
	require.Equal(t, uuid.UUID{}, second.InReplyTo)

	list, err := repo.ListByChannel(ctx, channel.ID, 10)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.False(t, list[1].CreatedAt.Before(list[0].CreatedAt), "messages come back in chronological order")

	_, err = repo.Get(ctx, testutil.RandomUUID())
	require.ErrorIs(t, err, ports.ErrMessageNotFound)
}

func TestCommandRepository(t *testing.T) {
	ctx := context.Background()
	repo := supabase.NewCommandRepository(testPool)
	company := seedCompany(t)
	agent := seedAgent(t, company.ID)

	// Global command: no company, no target agent.
	global, err := repo.Create(ctx, domain.Command{
		Kind:   domain.CommandKill,
		Status: domain.CommandPending,
		Reason: testutil.RandomText(),
	})
	require.NoError(t, err)
	require.Equal(t, uuid.UUID{}, global.CompanyID)
	require.Equal(t, uuid.UUID{}, global.TargetAgentID)
	require.Nil(t, global.AppliedAt)

	// Targeted command.
	targeted, err := repo.Create(ctx, domain.Command{
		CompanyID:     company.ID,
		TargetAgentID: agent.ID,
		Kind:          domain.CommandPause,
		Status:        domain.CommandPending,
	})
	require.NoError(t, err)
	require.Equal(t, agent.ID, targeted.TargetAgentID)

	now := time.Now().UTC()
	targeted.Status = domain.CommandApplied
	targeted.AppliedAt = &now
	applied, err := repo.Update(ctx, targeted)
	require.NoError(t, err)
	require.Equal(t, domain.CommandApplied, applied.Status)
	require.NotNil(t, applied.AppliedAt)

	pending, err := repo.ListByStatus(ctx, domain.CommandPending)
	require.NoError(t, err)
	require.NotEmpty(t, pending)
	for _, c := range pending {
		require.Equal(t, domain.CommandPending, c.Status)
	}

	_, err = repo.Get(ctx, testutil.RandomUUID())
	require.ErrorIs(t, err, ports.ErrCommandNotFound)
}

// TestMigrationDownDropsSchema is opt-in (set E2E_TEST_DOWN=1) because it drops
// every table and would break any test that runs after it.
func TestMigrationDownDropsSchema(t *testing.T) {
	if os.Getenv("E2E_TEST_DOWN") != "1" {
		t.Skip("set E2E_TEST_DOWN=1 to exercise 0001_init.down.sql")
	}
	if _, override := os.LookupEnv("E2E_DATABASE_URL"); override {
		t.Skip("refusing to run the down migration against an externally provided database")
	}

	ctx := context.Background()
	down, err := os.ReadFile("../migrations/0001_init.down.sql")
	require.NoError(t, err)

	_, err = testPool.Exec(ctx, string(down))
	require.NoError(t, err)

	var exists bool
	err = testPool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tasks')`,
	).Scan(&exists)
	require.NoError(t, err)
	require.False(t, exists, "0001_init.down.sql removed the tasks table")
}
