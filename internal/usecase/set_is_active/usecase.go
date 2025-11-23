package set_is_active

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/contract/repository/teams"
	"pr-reviewers-service/internal/usecase/contract/repository/users"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type usecase struct {
	repTeams teams.RepositoryTeams
	repUsers users.RepositoryUsers
	trm      trm.Manager
}

func NewUsecase(repTeams teams.RepositoryTeams, repUsers users.RepositoryUsers, trm trm.Manager) *usecase {
	return &usecase{
		repTeams: repTeams,
		repUsers: repUsers,
		trm:      trm,
	}
}

func (u *usecase) Run(ctx context.Context, req In) (*Out, error) {
	var result *Out
	var err error

	err = u.trm.Do(ctx, func(ctx context.Context) error {
		result, err = u.run(ctx, req)
		return err
	})

	return result, err
}

func (u *usecase) run(ctx context.Context, req In) (*Out, error) {
	slog.DebugContext(ctx, "Call GetUserByID", "user_id", req.UserID)

	existingUser, err := u.repUsers.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: user_id %s", usecase2.ErrUserNotFound, req.UserID))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetUser, req.UserID))
	}

	slog.DebugContext(ctx, "Call GetTeamByID", "team_id", existingUser.TeamID)
	team, err := u.repTeams.GetTeamByID(ctx, existingUser.TeamID)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: team_id %s", usecase2.ErrGetTeam, existingUser.TeamID))
	}

	if existingUser.IsActive == req.IsActive {
		slog.DebugContext(ctx, "User already has required is_active value",
			"user_id", req.UserID, "is_active", req.IsActive)
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrUserDontNeedChange))
	}

	slog.DebugContext(ctx, "Call UpdateUser", "user_id", req.UserID, "new_is_active", req.IsActive)

	userIn := users2.UserIn{
		ID:       existingUser.ID,
		Name:     existingUser.Name,
		IsActive: req.IsActive,
		TeamID:   existingUser.TeamID,
	}

	updatedUser, err := u.repUsers.UpdateUser(ctx, userIn)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrUpdateUser, req.UserID))
	}

	slog.DebugContext(ctx, "UseCase SetIsActive success", "user_id", req.UserID)
	return &Out{
		UserId:   updatedUser.ID,
		Username: updatedUser.Name,
		TeamName: team.Name,
		IsActive: updatedUser.IsActive,
	}, nil
}
