package pull_request_create

import (
	"context"

	"pr-reviewers-service/internal/usecase/pull_request_create"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=pull_request_create usecase
type usecase interface {
	Run(ctx context.Context, req pull_request_create.In) (*pull_request_create.Out, error)
}
