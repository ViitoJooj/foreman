package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func validCreateCompanyInput() service.CreateCompanyInput {
	return service.CreateCompanyInput{
		Name:      testutil.RandomCompanyName(),
		Slug:      testutil.RandomSlug(),
		RepoOwner: testutil.RandomRepoOwner(),
		RepoName:  testutil.RandomRepoName(),
	}
}

func TestCompanyCreatePersists(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockCompanyRepository(ctrl)

	in := validCreateCompanyInput()
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.Company) (domain.Company, error) {
			require.Equal(t, in.Name, c.Name)
			require.Equal(t, in.Slug, c.Slug)
			require.Equal(t, in.RepoOwner, c.RepoOwner)
			require.Equal(t, in.RepoName, c.RepoName)
			c.ID = testutil.RandomUUID()
			return c, nil
		})

	out, err := service.NewCompany(repo).Create(context.Background(), in)
	require.NoError(t, err)
	require.NotEqual(t, (domain.Company{}).ID, out.ID)
}

func TestCompanyCreateRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*service.CreateCompanyInput)
	}{
		{"missing name", func(in *service.CreateCompanyInput) { in.Name = "" }},
		{"missing slug", func(in *service.CreateCompanyInput) { in.Slug = "" }},
		{"missing repo owner", func(in *service.CreateCompanyInput) { in.RepoOwner = "" }},
		{"missing repo name", func(in *service.CreateCompanyInput) { in.RepoName = "" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := testutil.NewMockCompanyRepository(ctrl)

			in := validCreateCompanyInput()
			tc.mutate(&in)

			_, err := service.NewCompany(repo).Create(context.Background(), in)
			require.ErrorIs(t, err, service.ErrInvalidCompanyInput)
		})
	}
}

func TestCompanyGetAndListDelegate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockCompanyRepository(ctrl)
	uc := service.NewCompany(repo)

	id := testutil.RandomUUID()
	want := domain.Company{ID: id, Name: testutil.RandomCompanyName()}
	repo.EXPECT().Get(gomock.Any(), id).Return(want, nil)
	got, err := uc.Get(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, want, got)

	list := []domain.Company{want}
	repo.EXPECT().List(gomock.Any()).Return(list, nil)
	gotList, err := uc.List(context.Background())
	require.NoError(t, err)
	require.Equal(t, list, gotList)
}

func TestCompanyGetPropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockCompanyRepository(ctrl)

	repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(domain.Company{}, errors.New(testutil.RandomText()))
	_, err := service.NewCompany(repo).Get(context.Background(), testutil.RandomUUID())
	require.Error(t, err)
}
