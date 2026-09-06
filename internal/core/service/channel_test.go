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

func validCreateChannelInput() service.CreateChannelInput {
	return service.CreateChannelInput{
		CompanyID: testutil.RandomUUID(),
		Name:      testutil.RandomChannelName(),
	}
}

func TestChannelCreatePersists(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockChannelRepository(ctrl)

	in := validCreateChannelInput()
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ch domain.Channel) (domain.Channel, error) {
			require.Equal(t, in.CompanyID, ch.CompanyID)
			require.Equal(t, in.Name, ch.Name)
			ch.ID = testutil.RandomUUID()
			return ch, nil
		})

	_, err := service.NewChannel(repo).Create(context.Background(), in)
	require.NoError(t, err)
}

func TestChannelCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*service.CreateChannelInput)
	}{
		{"missing company", func(in *service.CreateChannelInput) { in.CompanyID = uuid.UUID{} }},
		{"missing name", func(in *service.CreateChannelInput) { in.Name = "" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := testutil.NewMockChannelRepository(ctrl)

			in := validCreateChannelInput()
			tc.mutate(&in)

			_, err := service.NewChannel(repo).Create(context.Background(), in)
			require.ErrorIs(t, err, service.ErrInvalidChannelInput)
		})
	}
}

func TestChannelGetAndListDelegate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockChannelRepository(ctrl)
	uc := service.NewChannel(repo)

	id := testutil.RandomUUID()
	want := domain.Channel{ID: id, Name: testutil.RandomChannelName()}
	repo.EXPECT().Get(gomock.Any(), id).Return(want, nil)
	got, err := uc.Get(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, want, got)

	companyID := testutil.RandomUUID()
	list := []domain.Channel{want}
	repo.EXPECT().ListByCompany(gomock.Any(), companyID).Return(list, nil)
	gotList, err := uc.List(context.Background(), companyID)
	require.NoError(t, err)
	require.Equal(t, list, gotList)
}
