package stats_pr_assignments

import (
	"context"

	"pr-reviewers-service/internal/usecase/stats_pr_assignments"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=stats_pr_assignments usecase
type usecase interface {
	Run(ctx context.Context, req stats_pr_assignments.In) (*stats_pr_assignments.Out, error)
}
