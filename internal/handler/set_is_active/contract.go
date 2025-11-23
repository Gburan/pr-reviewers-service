package set_is_active

import (
	"context"

	"pr-reviewers-service/internal/usecase/set_is_active"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=set_is_active usecase
type usecase interface {
	Run(ctx context.Context, req set_is_active.In) (*set_is_active.Out, error)
}
