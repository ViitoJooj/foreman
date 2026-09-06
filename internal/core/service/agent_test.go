package service_test

import (
	"context"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func validCreateAgentInput() service.CreateAgentInput {
	return service.CreateAgentInput{
		CompanyID: testutil.RandomUUID(),
		Name:      testutil.RandomAgentName(),
		Role:      testutil.RandomAgentRole(),
	}
}

func TestAgentCreatePersistsIdleAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockAgentRepository(ctrl)

	in := validCreateAgentInput()
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, a domain.Agent) (domain.Agent, error) {
			require.Equal(t, in.CompanyID, a.CompanyID)
			require.Equal(t, in.Name, a.Name)
			require.Equal(t, in.Role, a.Role)
			require.Equal(t, domain.AgentIdle, a.Status)
			a.ID = testutil.RandomUUID()
			return a, nil
		})

	_, err := service.NewAgent(repo).Create(context.Background(), in)
	require.NoError(t, err)
}

func TestAgentCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*service.CreateAgentInput)
	}{
		{"missing company", func(in *service.CreateAgentInput) { in.CompanyID = uuid.UUID{} }},
		{"missing name", func(in *service.CreateAgentInput) { in.Name = "" }},
		{"unknown role", func(in *service.CreateAgentInput) { in.Role = domain.AgentRole(testutil.RandomSlug()) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := testutil.NewMockAgentRepository(ctrl)

			in := validCreateAgentInput()
			tc.mutate(&in)

			_, err := service.NewAgent(repo).Create(context.Background(), in)
			require.ErrorIs(t, err, service.ErrInvalidAgentInput)
		})
	}
}

func TestAgentGetAndListDelegate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockAgentRepository(ctrl)
	uc := service.NewAgent(repo)

	id := testutil.RandomUUID()
	want := domain.Agent{ID: id, Name: testutil.RandomAgentName(), Role: testutil.RandomAgentRole()}
	repo.EXPECT().Get(gomock.Any(), id).Return(want, nil)
	got, err := uc.Get(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, want, got)

	companyID := testutil.RandomUUID()
	list := []domain.Agent{want}
	repo.EXPECT().ListByCompany(gomock.Any(), companyID).Return(list, nil)
	gotList, err := uc.List(context.Background(), companyID)
	require.NoError(t, err)
	require.Equal(t, list, gotList)
}
