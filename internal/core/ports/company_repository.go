package ports

import (
	"context"
	"errors"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

// ErrCompanyNotFound is returned when no company matches the given identifier.
var ErrCompanyNotFound = errors.New("company not found")

// CompanyRepository persists the repositories the harness maintains.
type CompanyRepository interface {
	Create(ctx context.Context, c domain.Company) (domain.Company, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Company, error)
	Update(ctx context.Context, c domain.Company) (domain.Company, error)
	List(ctx context.Context) ([]domain.Company, error)
}
