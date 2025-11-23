package pr_statuses

import (
	"context"

	"pr-reviewers-service/internal/infrastructure/repository/pr_statuses"

	"github.com/google/uuid"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=pr_statuses RepositoryPrStatuses
type RepositoryPrStatuses interface {
	SavePRStatus(ctx context.Context, status pr_statuses.PRStatusIn) (*pr_statuses.PRStatusOut, error)
	GetPRStatusByID(ctx context.Context, status pr_statuses.PRStatusIn) (*pr_statuses.PRStatusOut, error)
	GetPRStatusesByIDs(ctx context.Context, statusIDs []uuid.UUID) (*[]pr_statuses.PRStatusOut, error)
	UpdatePRStatusByID(ctx context.Context, status pr_statuses.PRStatusIn) (*pr_statuses.PRStatusOut, error)
}
