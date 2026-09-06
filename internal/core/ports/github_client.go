package ports

import "context"

// PullRequestInput carries the fields needed to open or update a pull request.
type PullRequestInput struct {
	Title string
	Body  string
	Head  string // source branch
	Base  string // target branch
}

// GitHubClient is a thin wrapper over the `gh` CLI. The repo argument is the
// "owner/name" identifier. Implementations shell out and never touch the store.
type GitHubClient interface {
	CreateBranch(ctx context.Context, repo, base, branch string) error
	OpenPR(ctx context.Context, repo string, in PullRequestInput) (number int, err error)
	UpdatePR(ctx context.Context, repo string, number int, in PullRequestInput) error
	MergePR(ctx context.Context, repo string, number int) error
	ClosePR(ctx context.Context, repo string, number int) error
	DeleteBranch(ctx context.Context, repo, branch string) error
}
