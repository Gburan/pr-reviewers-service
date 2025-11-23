package pull_request_merge

import (
	"context"

	"pr-reviewers-service/internal/usecase/pull_request_merge"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=pull_request_merge usecase
type usecase interface {
	Run(ctx context.Context, req pull_request_merge.In) (*pull_request_merge.Out, error)
}
