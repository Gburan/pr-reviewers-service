package add_team

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	teams2 "pr-reviewers-service/internal/infrastructure/repository/teams"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	"pr-reviewers-service/internal/logging"
	"pr-reviewers-service/internal/metrics"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/contract/repository/teams"
	"pr-reviewers-service/internal/usecase/contract/repository/users"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/google/uuid"
)

type usecase struct {
	repUsers users.RepositoryUsers
	repTeams teams.RepositoryTeams
	trm      trm.Manager
}

func Newusecase(repUsers users.RepositoryUsers, repTeams teams.RepositoryTeams, trm trm.Manager) *usecase {
	return &usecase{
		repUsers: repUsers,
		repTeams: repTeams,
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
	userIDSet := make(map[uuid.UUID]struct{})
	for _, member := range req.Members {
		if _, exist := userIDSet[member.UserID]; exist {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: duplicate user_id %s", usecase2.ErrDuplicateUsers, member.UserID))
		}
		userIDSet[member.UserID] = struct{}{}
	}

	slog.DebugContext(ctx, "Call GetTeamByName")
	existingTeam, err := u.repTeams.GetTeamByName(ctx, req.TeamName)
	if err != nil && !errors.Is(err, repository.ErrTeamNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetTeam, req.TeamName))
	}

	var teamID uuid.UUID
	var teamOut *teams2.TeamOut

	if existingTeam != nil {
		slog.DebugContext(ctx, "Found team")
		teamID = existingTeam.ID
		teamOut = existingTeam
	} else {
		slog.DebugContext(ctx, "Call SaveTeam")
		teamIn := teams2.TeamIn{
			ID:   uuid.New(),
			Name: req.TeamName,
		}
		teamOut, err = u.repTeams.SaveTeam(ctx, teamIn)
		if err != nil {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrSaveTeam, req.TeamName))
		}
		teamID = teamOut.ID
	}

	userIDs := make([]uuid.UUID, 0, len(req.Members))
	for _, member := range req.Members {
		userIDs = append(userIDs, member.UserID)
	}

	slog.DebugContext(ctx, "Call GetUsersByIDs")
	existingUsers, err := u.repUsers.GetUsersByIDs(ctx, userIDs)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrGetUsers))
	}

	existingUsersMap := make(map[uuid.UUID]users2.UserOut)
	if existingUsers != nil {
		for _, user := range *existingUsers {
			existingUsersMap[user.ID] = user
		}
	}

	var usersToUpdate []users2.UserIn
	var usersToCreate []users2.UserIn
	for _, member := range req.Members {
		userIn := users2.UserIn{
			ID:       member.UserID,
			Name:     member.Username,
			IsActive: member.IsActive,
			TeamID:   teamID,
		}
		if existingUser, exists := existingUsersMap[member.UserID]; exists {
			if !userNeedsUpdate(existingUser, userIn) {
				slog.DebugContext(ctx, "User dont need update", "user_id", member.UserID)
				continue
			}
			usersToUpdate = append(usersToUpdate, userIn)
		} else {
			usersToCreate = append(usersToCreate, userIn)
		}
	}

	var allUsers []users2.UserOut
	if len(usersToCreate) > 0 {
		slog.DebugContext(ctx, "Call SaveUsersBatch for new users")
		createdUsers, err := u.repUsers.SaveUsersBatch(ctx, usersToCreate)
		if err != nil {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrSaveUsersBatch))
		}
		allUsers = append(allUsers, *createdUsers...)
	}
	if len(usersToUpdate) > 0 {
		slog.DebugContext(ctx, "Call UpdateUsersBatch for existing users")
		updatedUsers, err := u.repUsers.UpdateUsersBatch(ctx, usersToUpdate)
		if err != nil {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrUpdateUsersBatch))
		}
		allUsers = append(allUsers, *updatedUsers...)
	}

	processedMembers := make([]TeamMembers, 0, len(allUsers))
	for _, user := range allUsers {
		processedMembers = append(processedMembers, TeamMembers{
			UserID:   user.ID,
			Username: user.Name,
			IsActive: user.IsActive,
		})
	}

	slog.DebugContext(ctx, "UseCase AddTeam success")
	if len(processedMembers) == 0 {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrNoUsersWereUpdatedAddedTeam))
	}

	metrics.IncCreatedTeams()
	metrics.IncCreatedUsers(len(processedMembers))
	return &Out{
		TeamName: teamOut.Name,
		Members:  processedMembers,
	}, nil
}

func userNeedsUpdate(existingUser users2.UserOut, newUser users2.UserIn) bool {
	return existingUser.Name != newUser.Name ||
		existingUser.IsActive != newUser.IsActive ||
		existingUser.TeamID != newUser.TeamID
}
