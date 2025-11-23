package get_team

import (
	"context"

	"pr-reviewers-service/internal/usecase/get_team"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=get_team usecase
type usecase interface {
	Run(ctx context.Context, req get_team.In) (*get_team.Out, error)
}
