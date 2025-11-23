package pr_reviewers

import (
	"context"

	"pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"

	"github.com/google/uuid"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=pr_reviewers RepositoryPrReviewers
type RepositoryPrReviewers interface {
	SavePRReviewer(ctx context.Context, reviewer pr_reviewers.PrReviewerIn) (*pr_reviewers.PrReviewerOut, error)
	GetPRReviewersByPRID(ctx context.Context, prID uuid.UUID) (*[]pr_reviewers.PrReviewerOut, error)
	GetPRReviewersByReviewerID(ctx context.Context, reviewerID uuid.UUID) (*[]pr_reviewers.PrReviewerOut, error)
	GetPRReviewersByReviewerIDs(ctx context.Context, reviewerIDs []uuid.UUID) (*[]pr_reviewers.PrReviewerOut, error)
	GetAllPRReviewers(ctx context.Context) (*[]pr_reviewers.PrReviewerOut, error)
	DeletePRReviewerByPRAndReviewer(ctx context.Context, prID, reviewerID uuid.UUID) error
}
