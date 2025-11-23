package add_team

import (
	"context"

	"pr-reviewers-service/internal/usecase/add_team"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=add_product usecase
type usecase interface {
	Run(ctx context.Context, req add_team.In) (*add_team.Out, error)
}
