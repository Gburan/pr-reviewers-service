package team_deactivate_users

import (
	"context"

	"pr-reviewers-service/internal/usecase/team_deactivate_users"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=team_deactivate_users usecase
type usecase interface {
	Run(ctx context.Context, req team_deactivate_users.In) (*team_deactivate_users.Out, error)
}
