package get_review

import (
	"context"

	"pr-reviewers-service/internal/usecase/get_review"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=get_review usecase
type usecase interface {
	Run(ctx context.Context, req get_review.In) (*get_review.Out, error)
}
