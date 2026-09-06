package github_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/github"
	"github.com/ViitoJooj/foreman/internal/core/ports"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

type call struct {
	name string
	args []string
}

// fakeRunner records calls and returns canned stdout keyed by a substring of the
// joined args (first match wins). Anything unmatched returns empty output.
type fakeRunner struct {
	calls   []call
	replies map[string]string
	failOn  string
}

func (f *fakeRunner) run(_ context.Context, name string, args ...string) ([]byte, error) {
	joined := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, call{name: name, args: args})
	if f.failOn != "" && strings.Contains(joined, f.failOn) {
		return nil, errors.New("gh failed: " + f.failOn)
	}
	for key, out := range f.replies {
		if strings.Contains(joined, key) {
			return []byte(out), nil
		}
	}
	return nil, nil
}

func TestCreateBranch(t *testing.T) {
	repo := testutil.RandomRepoOwner() + "/" + testutil.RandomRepoName()
	branch := testutil.RandomBranchSlug()
	fr := &fakeRunner{replies: map[string]string{"git/ref/heads/main": "abc123\n"}}

	c := github.New(github.WithRunner(fr.run))
	require.NoError(t, c.CreateBranch(context.Background(), repo, "main", branch))

	require.Len(t, fr.calls, 2)
	require.Equal(t, []string{"api", "repos/" + repo + "/git/ref/heads/main", "--jq", ".object.sha"}, fr.calls[0].args)
	require.Contains(t, fr.calls[1].args, "ref=refs/heads/"+branch)
	require.Contains(t, fr.calls[1].args, "sha=abc123")
}

func TestOpenPRParsesNumberFromURL(t *testing.T) {
	repo := "acme/site"
	fr := &fakeRunner{replies: map[string]string{"pr create": "https://github.com/acme/site/pull/57\n"}}

	c := github.New(github.WithRunner(fr.run))
	n, err := c.OpenPR(context.Background(), repo, ports.PullRequestInput{
		Title: "t", Body: "b", Head: "feature", Base: "main",
	})
	require.NoError(t, err)
	require.Equal(t, 57, n)
	require.Contains(t, fr.calls[0].args, "--head")
	require.Contains(t, fr.calls[0].args, "feature")
}

func TestOpenPRFallsBackToView(t *testing.T) {
	fr := &fakeRunner{replies: map[string]string{
		"pr create": "created\n", // no parseable number
		"pr view":   "88\n",
	}}
	c := github.New(github.WithRunner(fr.run))

	n, err := c.OpenPR(context.Background(), "acme/site", ports.PullRequestInput{Head: "x", Base: "main"})
	require.NoError(t, err)
	require.Equal(t, 88, n)
}

func TestMergePRUsesStrategy(t *testing.T) {
	fr := &fakeRunner{}
	c := github.New(github.WithRunner(fr.run), github.WithMergeStrategy("merge"))

	require.NoError(t, c.MergePR(context.Background(), "acme/site", 12))
	require.Equal(t, []string{"pr", "merge", "12", "--repo", "acme/site", "--merge", "--delete-branch"}, fr.calls[0].args)
}

func TestClosePRAndDeleteBranch(t *testing.T) {
	fr := &fakeRunner{}
	c := github.New(github.WithRunner(fr.run))

	require.NoError(t, c.ClosePR(context.Background(), "acme/site", 3))
	require.NoError(t, c.DeleteBranch(context.Background(), "acme/site", "autodev/x/y"))

	require.Equal(t, []string{"pr", "close", "3", "--repo", "acme/site"}, fr.calls[0].args)
	require.Equal(t, []string{"api", "-X", "DELETE", "repos/acme/site/git/refs/heads/autodev/x/y"}, fr.calls[1].args)
}

func TestUpdatePR(t *testing.T) {
	fr := &fakeRunner{}
	c := github.New(github.WithRunner(fr.run))

	require.NoError(t, c.UpdatePR(context.Background(), "acme/site", 9, ports.PullRequestInput{Title: "new", Body: "body"}))
	require.Equal(t, []string{"pr", "edit", "9", "--repo", "acme/site", "--title", "new", "--body", "body"}, fr.calls[0].args)
}

func TestCreateBranchSurfacesFailure(t *testing.T) {
	fr := &fakeRunner{failOn: "git/ref/heads/main"}
	c := github.New(github.WithRunner(fr.run))

	err := c.CreateBranch(context.Background(), "acme/site", "main", "x")
	require.ErrorContains(t, err, "resolve base")
}
