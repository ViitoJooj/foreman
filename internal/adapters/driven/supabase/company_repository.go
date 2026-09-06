package supabase

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

var _ ports.CompanyRepository = (*CompanyRepository)(nil)

// CompanyRepository is the Supabase-backed implementation of ports.CompanyRepository.
type CompanyRepository struct {
	pool *pgxpool.Pool
}

// NewCompanyRepository builds a CompanyRepository bound to the given pool.
func NewCompanyRepository(pool *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{pool: pool}
}

const companyColumns = `id, name, slug, repo_owner, repo_name, created_at, updated_at`

func scanCompany(row pgx.Row) (domain.Company, error) {
	var c domain.Company
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.RepoOwner, &c.RepoName, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

// Create inserts a company and returns it with the database-assigned id and timestamps.
func (r *CompanyRepository) Create(ctx context.Context, c domain.Company) (domain.Company, error) {
	const q = `
		INSERT INTO companies (name, slug, repo_owner, repo_name)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + companyColumns
	created, err := scanCompany(r.pool.QueryRow(ctx, q, c.Name, c.Slug, c.RepoOwner, c.RepoName))
	if err != nil {
		return domain.Company{}, fmt.Errorf("create company: %w", err)
	}
	return created, nil
}

// Get returns the company with the given id, or ports.ErrCompanyNotFound.
func (r *CompanyRepository) Get(ctx context.Context, id uuid.UUID) (domain.Company, error) {
	c, err := scanCompany(r.pool.QueryRow(ctx, `SELECT `+companyColumns+` FROM companies WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Company{}, ports.ErrCompanyNotFound
	}
	if err != nil {
		return domain.Company{}, fmt.Errorf("get company: %w", err)
	}
	return c, nil
}

// Update overwrites the mutable fields of an existing company.
func (r *CompanyRepository) Update(ctx context.Context, c domain.Company) (domain.Company, error) {
	const q = `
		UPDATE companies SET name = $2, slug = $3, repo_owner = $4, repo_name = $5
		WHERE id = $1
		RETURNING ` + companyColumns
	updated, err := scanCompany(r.pool.QueryRow(ctx, q, c.ID, c.Name, c.Slug, c.RepoOwner, c.RepoName))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Company{}, ports.ErrCompanyNotFound
	}
	if err != nil {
		return domain.Company{}, fmt.Errorf("update company: %w", err)
	}
	return updated, nil
}

// List returns every company, oldest first.
func (r *CompanyRepository) List(ctx context.Context) ([]domain.Company, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+companyColumns+` FROM companies ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	defer rows.Close()

	var companies []domain.Company
	for rows.Next() {
		c, err := scanCompany(rows)
		if err != nil {
			return nil, fmt.Errorf("scan company: %w", err)
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}
