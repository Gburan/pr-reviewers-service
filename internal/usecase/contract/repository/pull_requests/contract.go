package pull_requests

import (
	"context"

	"pr-reviewers-service/internal/infrastructure/repository/pull_requests"

	"github.com/google/uuid"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=pull_requests RepositoryPullRequests
type RepositoryPullRequests interface {
	SavePullRequest(ctx context.Context, pr pull_requests.PullRequestIn) (*pull_requests.PullRequestOut, error)
	GetPullRequestByID(ctx context.Context, prID uuid.UUID) (*pull_requests.PullRequestOut, error)
	GetPullRequestsByPrIDs(ctx context.Context, prIDs []uuid.UUID) (*[]pull_requests.PullRequestOut, error)
	MarkPullRequestMergedByID(ctx context.Context, prID uuid.UUID) (*pull_requests.PullRequestOut, error)
}
