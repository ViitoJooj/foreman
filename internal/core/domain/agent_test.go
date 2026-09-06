package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

// TestAgentEnumValues pins the wire values of the agent enums, which must match
// the agent_role and agent_status Postgres enums in migrations/0001_init.up.sql.
func TestAgentEnumValues(t *testing.T) {
	roles := map[domain.AgentRole]string{
		domain.RoleTaskCreator: "task_creator",
		domain.RoleCoder:       "coder",
		domain.RoleTester:      "tester",
		domain.RolePRReviewer:  "pr_reviewer",
	}
	for role, want := range roles {
		require.Equal(t, want, string(role))
	}

	statuses := map[domain.AgentStatus]string{
		domain.AgentIdle:    "idle",
		domain.AgentWorking: "working",
		domain.AgentPaused:  "paused",
	}
	for status, want := range statuses {
		require.Equal(t, want, string(status))
	}
}

func TestAgentHoldsAssignedFields(t *testing.T) {
	id := testutil.RandomUUID()
	companyID := testutil.RandomUUID()
	name := testutil.RandomAgentName()
	role := testutil.RandomAgentRole()
	status := testutil.RandomAgentStatus()

	a := domain.Agent{ID: id, CompanyID: companyID, Name: name, Role: role, Status: status}

	require.Equal(t, id, a.ID)
	require.Equal(t, companyID, a.CompanyID)
	require.Equal(t, name, a.Name)
	require.Equal(t, role, a.Role)
	require.Equal(t, status, a.Status)
}
