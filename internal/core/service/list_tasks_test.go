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

func taskWithState(companyID uuid.UUID, state domain.TaskState) domain.Task {
	return domain.Task{
		ID:        testutil.RandomUUID(),
		CompanyID: companyID,
		Title:     testutil.RandomTaskTitle(),
		State:     state,
		Risk:      testutil.RandomRiskLevel(),
	}
}

func TestListTasksExecuteByCompany(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockTaskRepository(ctrl)

	companyID := testutil.RandomUUID()
	want := []domain.Task{
		taskWithState(companyID, domain.TaskCoding),
		taskWithState(companyID, domain.TaskMerged),
	}
	repo.EXPECT().ListByCompany(gomock.Any(), companyID).Return(want, nil)

	got, err := service.NewListTasks(repo).Execute(context.Background(), service.ListTasksFilter{CompanyID: companyID})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestListTasksExecuteByState(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockTaskRepository(ctrl)

	state := testutil.RandomTaskState()
	want := []domain.Task{taskWithState(testutil.RandomUUID(), state)}
	repo.EXPECT().ListByState(gomock.Any(), state).Return(want, nil)

	got, err := service.NewListTasks(repo).Execute(context.Background(), service.ListTasksFilter{State: state})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestListTasksExecuteByCompanyAndStateFiltersInMemory(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockTaskRepository(ctrl)

	companyID := testutil.RandomUUID()
	wanted := taskWithState(companyID, domain.TaskTesting)
	other := taskWithState(companyID, domain.TaskCoding)
	repo.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Task{wanted, other}, nil)

	got, err := service.NewListTasks(repo).Execute(context.Background(), service.ListTasksFilter{
		CompanyID: companyID,
		State:     domain.TaskTesting,
	})
	require.NoError(t, err)
	require.Equal(t, []domain.Task{wanted}, got)
}

func TestListTasksExecuteRequiresAFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockTaskRepository(ctrl)
	// No repo call expected.

	_, err := service.NewListTasks(repo).Execute(context.Background(), service.ListTasksFilter{})
	require.ErrorIs(t, err, service.ErrTaskFilterRequired)
}

func TestListTasksExecutePropagatesRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockTaskRepository(ctrl)

	wantErr := errors.New(testutil.RandomText())
	repo.EXPECT().ListByState(gomock.Any(), gomock.Any()).Return(nil, wantErr)

	_, err := service.NewListTasks(repo).Execute(context.Background(), service.ListTasksFilter{State: testutil.RandomTaskState()})
	require.ErrorIs(t, err, wantErr)
}
