package http

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestNewTaskResponseMapsFields(t *testing.T) {
	task := domain.Task{
		ID:          testutil.RandomUUID(),
		CompanyID:   testutil.RandomUUID(),
		Title:       testutil.RandomTaskTitle(),
		Description: testutil.RandomText(),
		State:       domain.TaskCoding,
		Risk:        domain.RiskLow,
		Branch:      testutil.RandomBranchSlug(),
		PRNumber:    testutil.RandomIntBetween(1, 999),
		Retries:     1,
		MaxRetries:  3,
		AssigneeID:  testutil.RandomUUID(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	got := newTaskResponse(task)

	require.Equal(t, task.ID.String(), got.ID)
	require.Equal(t, task.CompanyID.String(), got.CompanyID)
	require.Equal(t, task.Branch, got.Branch)
	require.Equal(t, task.PRNumber, got.PRNumber)
	require.Equal(t, "coding", got.State)
	require.Equal(t, "low", got.Risk)
	require.Equal(t, task.AssigneeID.String(), got.AssigneeID)
}

func TestNewTaskResponseOmitsZeroAssignee(t *testing.T) {
	got := newTaskResponse(domain.Task{ID: testutil.RandomUUID(), CompanyID: testutil.RandomUUID()})
	require.Empty(t, got.AssigneeID)
}

func TestNewMessageResponseMapsFields(t *testing.T) {
	msg := domain.Message{
		ID:               testutil.RandomUUID(),
		ChannelID:        testutil.RandomUUID(),
		Type:             domain.MessageRequest,
		FromAgentID:      testutil.RandomUUID(),
		ToAgentID:        testutil.RandomUUID(),
		InReplyTo:        testutil.RandomUUID(),
		RequiresResponse: true,
		Body:             testutil.RandomText(),
		Payload:          testutil.RandomJSONPayload(),
		CreatedAt:        time.Now(),
	}

	got := newMessageResponse(msg)

	require.Equal(t, msg.ID.String(), got.ID)
	require.Equal(t, msg.ChannelID.String(), got.ChannelID)
	require.Equal(t, "request", got.Type)
	require.Equal(t, msg.FromAgentID.String(), got.FromAgentID)
	require.Equal(t, msg.ToAgentID.String(), got.ToAgentID)
	require.Equal(t, msg.InReplyTo.String(), got.InReplyTo)
	require.True(t, got.RequiresResponse)
	require.JSONEq(t, string(msg.Payload), string(got.Payload))
}

func TestNewMessageResponseOmitsZeroOptionalIDs(t *testing.T) {
	got := newMessageResponse(domain.Message{
		ID:          testutil.RandomUUID(),
		ChannelID:   testutil.RandomUUID(),
		Type:        domain.MessageStatus,
		FromAgentID: testutil.RandomUUID(),
	})
	require.Empty(t, got.ToAgentID)
	require.Empty(t, got.InReplyTo)
}

func TestOptionalUUID(t *testing.T) {
	require.Empty(t, optionalUUID(uuid.UUID{}))

	id := testutil.RandomUUID()
	require.Equal(t, id.String(), optionalUUID(id))
}

func TestParseOptionalUUID(t *testing.T) {
	t.Run("empty string yields the zero uuid", func(t *testing.T) {
		got, err := parseOptionalUUID("")
		require.NoError(t, err)
		require.Equal(t, uuid.UUID{}, got)
	})

	t.Run("valid string parses", func(t *testing.T) {
		id := testutil.RandomUUID()
		got, err := parseOptionalUUID(id.String())
		require.NoError(t, err)
		require.Equal(t, id, got)
	})

	t.Run("invalid string errors", func(t *testing.T) {
		_, err := parseOptionalUUID(testutil.RandomSlug())
		require.Error(t, err)
	})
}
