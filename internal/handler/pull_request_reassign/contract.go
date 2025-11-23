package pull_request_reassign

import (
	"context"

	"pr-reviewers-service/internal/usecase/pull_request_reassign"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=pull_request_reassign usecase
type usecase interface {
	Run(ctx context.Context, req pull_request_reassign.In) (*pull_request_reassign.Out, error)
}
