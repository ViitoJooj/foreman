package service_test

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestCommandIssueRecordsPending(t *testing.T) {
	tests := []struct {
		name  string
		kind  domain.CommandKind
		agent uuid.UUID
	}{
		{"kill is global", domain.CommandKill, uuid.UUID{}},
		{"panic is global", domain.CommandPanic, uuid.UUID{}},
		{"pause targets an agent", domain.CommandPause, testutil.RandomUUID()},
		{"resume targets an agent", domain.CommandResume, testutil.RandomUUID()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := testutil.NewMockCommandRepository(ctrl)

			in := service.IssueCommandInput{
				Kind:          tc.kind,
				CompanyID:     testutil.RandomUUID(),
				TargetAgentID: tc.agent,
				Reason:        testutil.RandomText(),
			}
			repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, c domain.Command) (domain.Command, error) {
					require.Equal(t, tc.kind, c.Kind)
					require.Equal(t, in.CompanyID, c.CompanyID)
					require.Equal(t, tc.agent, c.TargetAgentID)
					require.Equal(t, domain.CommandPending, c.Status)
					c.ID = testutil.RandomUUID()
					return c, nil
				})

			_, err := service.NewCommand(repo).Issue(context.Background(), in)
			require.NoError(t, err)
		})
	}
}

func TestCommandIssueRejectsInconsistentTarget(t *testing.T) {
	tests := []struct {
		name  string
		kind  domain.CommandKind
		agent uuid.UUID
	}{
		{"kill with an agent", domain.CommandKill, testutil.RandomUUID()},
		{"panic with an agent", domain.CommandPanic, testutil.RandomUUID()},
		{"pause without an agent", domain.CommandPause, uuid.UUID{}},
		{"resume without an agent", domain.CommandResume, uuid.UUID{}},
		{"unknown kind", domain.CommandKind(testutil.RandomSlug()), uuid.UUID{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := testutil.NewMockCommandRepository(ctrl)
			// No repo call expected.

			_, err := service.NewCommand(repo).Issue(context.Background(), service.IssueCommandInput{
				Kind:          tc.kind,
				TargetAgentID: tc.agent,
			})
			require.ErrorIs(t, err, service.ErrInvalidCommandInput)
		})
	}
}

func TestCommandGetAndListDelegate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockCommandRepository(ctrl)
	uc := service.NewCommand(repo)

	id := testutil.RandomUUID()
	want := domain.Command{ID: id, Kind: domain.CommandKill, Status: domain.CommandPending}
	repo.EXPECT().Get(gomock.Any(), id).Return(want, nil)
	got, err := uc.Get(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, want, got)

	list := []domain.Command{want}
	repo.EXPECT().ListByStatus(gomock.Any(), domain.CommandPending).Return(list, nil)
	gotList, err := uc.List(context.Background(), domain.CommandPending)
	require.NoError(t, err)
	require.Equal(t, list, gotList)
}

func TestCommandListPropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockCommandRepository(ctrl)

	repo.EXPECT().ListByStatus(gomock.Any(), gomock.Any()).Return(nil, errors.New(testutil.RandomText()))
	_, err := service.NewCommand(repo).List(context.Background(), domain.CommandApplied)
	require.Error(t, err)
}
