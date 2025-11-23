package get_team

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/contract/repository/teams"
	users2 "pr-reviewers-service/internal/usecase/contract/repository/users"
)

type usecase struct {
	repTeams teams.RepositoryTeams
	repUsers users2.RepositoryUsers
}

func NewUsecase(repTeams teams.RepositoryTeams, repUsers users2.RepositoryUsers) *usecase {
	return &usecase{
		repTeams: repTeams,
		repUsers: repUsers,
	}
}

func (u *usecase) Run(ctx context.Context, req In) (*Out, error) {
	slog.DebugContext(ctx, "Call GetTeamByName")
	team, err := u.repTeams.GetTeamByName(ctx, req.TeamName)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: team %s", usecase2.ErrTeamNotFound, req.TeamName))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetTeam, req.TeamName))
	}

	slog.DebugContext(ctx, "Call GetUsersByTeamID")
	users, err := u.repUsers.GetUsersByTeamID(ctx, team.ID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: team_id %s", usecase2.ErrGetUsers, team.ID))
	}

	members := make([]TeamMembers, 0)
	if users != nil {
		for _, user := range *users {
			members = append(members, TeamMembers{
				UserID:   user.ID,
				Username: user.Name,
				IsActive: user.IsActive,
			})
		}
	}

	slog.DebugContext(ctx, "UseCase GetTeam success")
	return &Out{
		TeamName: team.Name,
		Members:  members,
	}, nil
}
