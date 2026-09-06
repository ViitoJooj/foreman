package service

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// ErrInvalidCompanyInput is returned when Company.Create is given missing fields.
var ErrInvalidCompanyInput = errors.New("invalid company input")

// Company groups the company use cases. Reads are thin passthroughs; Create
// validates before delegating.
type Company struct {
	repo ports.CompanyRepository
}

// NewCompany wires the use cases to a company repository.
func NewCompany(repo ports.CompanyRepository) *Company {
	return &Company{repo: repo}
}

// CreateCompanyInput carries the fields required to register a company.
type CreateCompanyInput struct {
	Name      string
	Slug      string
	RepoOwner string
	RepoName  string
}

// Create validates the input and persists the new company.
func (uc *Company) Create(ctx context.Context, in CreateCompanyInput) (domain.Company, error) {
	if in.Name == "" || in.Slug == "" || in.RepoOwner == "" || in.RepoName == "" {
		return domain.Company{}, fmt.Errorf("%w: name, slug, repo_owner and repo_name are required", ErrInvalidCompanyInput)
	}
	return uc.repo.Create(ctx, domain.Company{
		Name:      in.Name,
		Slug:      in.Slug,
		RepoOwner: in.RepoOwner,
		RepoName:  in.RepoName,
	})
}

// Get returns the company with the given id.
func (uc *Company) Get(ctx context.Context, id uuid.UUID) (domain.Company, error) {
	return uc.repo.Get(ctx, id)
}

// List returns every company, oldest first.
func (uc *Company) List(ctx context.Context) ([]domain.Company, error) {
	return uc.repo.List(ctx)
}
