package teams

import (
	"context"

	"pr-reviewers-service/internal/infrastructure/repository/teams"

	"github.com/google/uuid"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=teams RepositoryTeams
type RepositoryTeams interface {
	SaveTeam(ctx context.Context, team teams.TeamIn) (*teams.TeamOut, error)
	GetTeamByID(ctx context.Context, team uuid.UUID) (*teams.TeamOut, error)
	GetTeamByName(ctx context.Context, name string) (*teams.TeamOut, error)
}
