package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestCompanyRepoFullName(t *testing.T) {
	owner := testutil.RandomRepoOwner()
	name := testutil.RandomRepoName()

	c := domain.Company{
		ID:        testutil.RandomUUID(),
		Name:      testutil.RandomCompanyName(),
		Slug:      testutil.RandomSlug(),
		RepoOwner: owner,
		RepoName:  name,
	}

	require.Equal(t, owner+"/"+name, c.RepoFullName())
}

func TestCompanyRepoFullNameZeroValue(t *testing.T) {
	require.Equal(t, "/", domain.Company{}.RepoFullName())
}
