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

func validCreateTaskInput() service.CreateTaskInput {
	return service.CreateTaskInput{
		CompanyID:   testutil.RandomUUID(),
		Title:       testutil.RandomTaskTitle(),
		Description: testutil.RandomText(),
		Risk:        testutil.RandomRiskLevel(),
	}
}

func TestCreateTaskExecutePersistsCreatedTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockTaskRepository(ctrl)

	in := validCreateTaskInput()

	var got domain.Task
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, task domain.Task) (domain.Task, error) {
			got = task
			task.ID = testutil.RandomUUID()
			return task, nil
		})

	out, err := service.NewCreateTask(repo).Execute(context.Background(), in)
	require.NoError(t, err)

	require.Equal(t, in.CompanyID, got.CompanyID)
	require.Equal(t, in.Title, got.Title)
	require.Equal(t, in.Description, got.Description)
	require.Equal(t, in.Risk, got.Risk)
	require.Equal(t, domain.TaskCreated, got.State)
	require.Equal(t, 3, got.MaxRetries, "non-positive max retries falls back to the default")
	require.NotEqual(t, (domain.Task{}).ID, out.ID)
}

func TestCreateTaskExecuteKeepsPositiveMaxRetries(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockTaskRepository(ctrl)

	in := validCreateTaskInput()
	in.MaxRetries = testutil.RandomIntBetween(1, 9)

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, task domain.Task) (domain.Task, error) {
			require.Equal(t, in.MaxRetries, task.MaxRetries)
			return task, nil
		})

	_, err := service.NewCreateTask(repo).Execute(context.Background(), in)
	require.NoError(t, err)
}

func TestCreateTaskExecuteRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*service.CreateTaskInput)
	}{
		{"missing company", func(in *service.CreateTaskInput) { in.CompanyID = uuid.UUID{} }},
		{"missing title", func(in *service.CreateTaskInput) { in.Title = "" }},
		{"unknown risk", func(in *service.CreateTaskInput) { in.Risk = domain.RiskLevel(testutil.RandomSlug()) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := testutil.NewMockTaskRepository(ctrl)
			// No repo call expected.

			in := validCreateTaskInput()
			tc.mutate(&in)

			_, err := service.NewCreateTask(repo).Execute(context.Background(), in)
			require.ErrorIs(t, err, service.ErrInvalidTaskInput)
		})
	}
}

func TestCreateTaskExecutePropagatesRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockTaskRepository(ctrl)

	wantErr := errors.New(testutil.RandomText())
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.Task{}, wantErr)

	_, err := service.NewCreateTask(repo).Execute(context.Background(), validCreateTaskInput())
	require.ErrorIs(t, err, wantErr)
}
