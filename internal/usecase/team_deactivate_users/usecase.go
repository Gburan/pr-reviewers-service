package team_deactivate_users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	pr_reviewers2 "pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/contract/randomizer"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_statuses"
	"pr-reviewers-service/internal/usecase/contract/repository/pull_requests"
	"pr-reviewers-service/internal/usecase/contract/repository/teams"
	"pr-reviewers-service/internal/usecase/contract/repository/users"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/google/uuid"
)

type usecase struct {
	repTeams        teams.RepositoryTeams
	repUsers        users.RepositoryUsers
	repPullRequests pull_requests.RepositoryPullRequests
	repPRReviewers  pr_reviewers.RepositoryPrReviewers
	repPRStatuses   pr_statuses.RepositoryPrStatuses
	randomizer      randomizer.Randomizer
	trm             trm.Manager
}

func NewUsecase(
	repTeams teams.RepositoryTeams,
	repUsers users.RepositoryUsers,
	repPullRequests pull_requests.RepositoryPullRequests,
	repPRReviewers pr_reviewers.RepositoryPrReviewers,
	repPRStatuses pr_statuses.RepositoryPrStatuses,
	randomizer randomizer.Randomizer,
	trm trm.Manager,
) *usecase {
	return &usecase{
		repTeams:        repTeams,
		repUsers:        repUsers,
		repPullRequests: repPullRequests,
		repPRReviewers:  repPRReviewers,
		repPRStatuses:   repPRStatuses,
		randomizer:      randomizer,
		trm:             trm,
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
	slog.DebugContext(ctx, "Get team by name", "team_name", req.TeamName)
	team, err := u.repTeams.GetTeamByName(ctx, req.TeamName)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: team %s", usecase2.ErrTeamNotFound, req.TeamName))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetTeam, req.TeamName))
	}

	slog.DebugContext(ctx, "Get users by IDs", "user_ids_count", len(req.UserIDs))
	existingUsers, err := u.repUsers.GetUsersByIDs(ctx, req.UserIDs)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrGetUsers))
	}
	if existingUsers == nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrUsersByIDsNotFound))
	}
	userIDs := make(map[uuid.UUID]struct{})
	for _, user := range *existingUsers {
		if user.TeamID != team.ID {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: user %s not in team %s", usecase2.ErrUserNotBelongsToTeam, user.ID, req.TeamName))
		}
		userIDs[user.ID] = struct{}{}
	}
	for _, userID := range req.UserIDs {
		if _, exist := userIDs[userID]; !exist {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: user %s", usecase2.ErrUserNotFound, userID))
		}
	}

	slog.DebugContext(ctx, "Find affected PRs for deactivation")
	prsToAffect, err := u.findPRsToAffect(ctx, req.UserIDs)
	if err != nil {
		return nil, err
	}
	if len(prsToAffect) == 0 {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrNoPRsToAffect))
	}

	slog.DebugContext(ctx, "Deactivate users", "users_count", len(req.UserIDs))
	var usersToUpdate []users2.UserIn
	for _, user := range *existingUsers {
		if user.IsActive {
			usersToUpdate = append(usersToUpdate, users2.UserIn{
				ID:       user.ID,
				Name:     user.Name,
				IsActive: false,
				TeamID:   user.TeamID,
			})
		}
	}

	if len(usersToUpdate) > 0 {
		_, err = u.repUsers.UpdateUsersBatch(ctx, usersToUpdate)
		if err != nil {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrUpdateUser))
		}
	}

	slog.DebugContext(ctx, "Reassign PR reviewers for affected PRs", "affected_prs_count", len(prsToAffect))
	reassignedPRs, err := u.reassignPRReviewers(ctx, prsToAffect, req.UserIDs)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Get updated team members")
	updatedUsers, err := u.repUsers.GetUsersByTeamID(ctx, team.ID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: team_id %s", usecase2.ErrGetUsers, team.ID))
	}

	members := make([]TeamMember, 0)
	if updatedUsers != nil {
		for _, user := range *updatedUsers {
			members = append(members, TeamMember{
				UserID:   user.ID,
				Username: user.Name,
				IsActive: user.IsActive,
			})
		}
	}

	slog.DebugContext(ctx, "UseCase DeactivateTeamUsers success",
		"deactivated_users", len(usersToUpdate),
		"affected_prs", len(reassignedPRs))
	return &Out{
		Team: Team{
			TeamName: team.Name,
			Members:  members,
		},
		AffectedPullRequests: reassignedPRs,
	}, nil
}

func (u *usecase) findPRsToAffect(ctx context.Context, userIDs []uuid.UUID) ([]PullRequestShort, error) {
	allReviewers, err := u.repPRReviewers.GetPRReviewersByReviewerIDs(ctx, userIDs)
	if err != nil && !errors.Is(err, repository.ErrPRReviewerNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrGetPRReviewers))
	}
	if allReviewers == nil || len(*allReviewers) == 0 {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrNoUsersAssignedToPRs))
	}

	prIDs := make([]uuid.UUID, 0, len(*allReviewers))
	prIDSet := make(map[uuid.UUID]struct{})
	for _, reviewer := range *allReviewers {
		if _, exist := prIDSet[reviewer.PRID]; !exist {
			prIDSet[reviewer.PRID] = struct{}{}
			prIDs = append(prIDs, reviewer.PRID)
		}
	}
	prs, err := u.repPullRequests.GetPullRequestsByPrIDs(ctx, prIDs)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrGetPullRequest))
	}
	statusIDs := make([]uuid.UUID, 0, len(*prs))
	for _, pr := range *prs {
		statusIDs = append(statusIDs, pr.StatusID)
	}
	prStatuses, err := u.repPRStatuses.GetPRStatusesByIDs(ctx, statusIDs)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrGetPRStatus))
	}
	statusMap := make(map[uuid.UUID]string)
	if prStatuses != nil {
		for _, status := range *prStatuses {
			statusMap[status.ID] = status.Status
		}
	}

	affectedPRs := make([]PullRequestShort, 0)
	for _, pr := range *prs {
		status, exists := statusMap[pr.StatusID]
		if !exists {
			slog.DebugContext(ctx, "Status not found for PR, skipping", "pr_id", pr.ID, "status_id", pr.StatusID)
			continue
		}

		if status == usecase2.OpenStatusValue {
			affectedPRs = append(affectedPRs, PullRequestShort{
				PullRequestID:   pr.ID,
				PullRequestName: pr.Name,
				AuthorID:        pr.AuthorID,
				Status:          status,
			})
		}
	}

	slog.DebugContext(ctx, "Found affected PRs", "total_prs", len(*prs), "open_prs", len(affectedPRs))
	return affectedPRs, nil
}

func (u *usecase) reassignPRReviewers(ctx context.Context, affectedPRs []PullRequestShort, deactivatedUserIDs []uuid.UUID) ([]PullRequestShort, error) {
	usersToDeactivateMap := make(map[uuid.UUID]struct{})
	for _, userID := range deactivatedUserIDs {
		usersToDeactivateMap[userID] = struct{}{}
	}

	reassignedPRs := make([]PullRequestShort, 0, len(affectedPRs))
	for _, pr := range affectedPRs {
		slog.DebugContext(ctx, "Processing PR for reviewer reassignment", "pr_id", pr.PullRequestID)

		currentReviewers, err := u.repPRReviewers.GetPRReviewersByPRID(ctx, pr.PullRequestID)
		if err != nil && !errors.Is(err, repository.ErrPRReviewerNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: pr_id %s", usecase2.ErrGetPRReviewers, pr.PullRequestID))
		}
		if currentReviewers == nil || len(*currentReviewers) == 0 {
			reassignedPRs = append(reassignedPRs, pr)
			continue
		}

		usersToDeactivate := make([]uuid.UUID, 0)
		for _, reviewer := range *currentReviewers {
			if _, isDeactivated := usersToDeactivateMap[reviewer.ReviewerID]; isDeactivated {
				usersToDeactivate = append(usersToDeactivate, reviewer.ReviewerID)
			}
		}
		if len(usersToDeactivate) == 0 {
			continue
		}

		prInfo, err := u.repPullRequests.GetPullRequestByID(ctx, pr.PullRequestID)
		if err != nil {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetPullRequest, pr.PullRequestID))
		}
		author, err := u.repUsers.GetUserByID(ctx, prInfo.AuthorID)
		if err != nil {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetUser, prInfo.AuthorID))
		}
		teamMembers, err := u.repUsers.GetActiveUsersByTeamID(ctx, author.TeamID)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: team_id %s", usecase2.ErrGetUsers, author.TeamID))
		}
		availableReviewers := u.getAvailableReviewersFromTeam(*teamMembers, author.ID, *currentReviewers)
		if len(availableReviewers) == 0 {
			slog.WarnContext(ctx, "No available reviewers found for PR", "pr_id", pr.PullRequestID, "team_id", author.TeamID)
			for _, reviewerID := range usersToDeactivate {
				err := u.repPRReviewers.DeletePRReviewerByPRAndReviewer(ctx, pr.PullRequestID, reviewerID)
				if err != nil {
					return nil, logging.WrapError(ctx, fmt.Errorf("%w: reviewer_id %s", usecase2.ErrRemoveReviewer, reviewerID))
				}
			}
			reassignedPRs = append(reassignedPRs, pr)
			continue
		}
		newReviewers := u.selectRandomReviewers(availableReviewers, len(usersToDeactivate))

		for _, reviewer := range newReviewers {
			reviewerIn := pr_reviewers2.PrReviewerIn{
				PrID:       pr.PullRequestID,
				ReviewerID: reviewer.ID,
			}
			_, err = u.repPRReviewers.SavePRReviewer(ctx, reviewerIn)
			if err != nil {
				return nil, logging.WrapError(ctx, fmt.Errorf("%w: reviewer %s", usecase2.ErrAssignReviewer, reviewer.ID))
			}
			slog.DebugContext(ctx, "Assigned new reviewer", "pr_id", pr.PullRequestID, "reviewer_id", reviewer.ID)
		}

		for _, reviewerID := range usersToDeactivate {
			err := u.repPRReviewers.DeletePRReviewerByPRAndReviewer(ctx, pr.PullRequestID, reviewerID)
			if err != nil {
				return nil, logging.WrapError(ctx, fmt.Errorf("%w: reviewer_id %s", usecase2.ErrRemoveReviewer, reviewerID))
			}
		}

		reassignedPRs = append(reassignedPRs, pr)
		slog.DebugContext(ctx, "PR reassignment completed",
			"pr_id", pr.PullRequestID,
			"removed_reviewers", len(usersToDeactivate),
			"added_reviewers", func() int {
				if teamMembers == nil || len(availableReviewers) == 0 {
					return 0
				}
				return min(len(usersToDeactivate), len(availableReviewers))
			}())
	}

	return reassignedPRs, nil
}

func (u *usecase) getAvailableReviewersFromTeam(
	teamMembers []users2.UserOut,
	authorID uuid.UUID,
	currentReviewers []pr_reviewers2.PrReviewerOut,
) []users2.UserOut {
	var available []users2.UserOut

	currentReviewerMap := make(map[uuid.UUID]bool)
	for _, reviewer := range currentReviewers {
		currentReviewerMap[reviewer.ReviewerID] = true
	}

	for _, member := range teamMembers {
		if member.ID == authorID || !member.IsActive || currentReviewerMap[member.ID] {
			continue
		}
		available = append(available, member)
	}
	return available
}

func (u *usecase) selectRandomReviewers(available []users2.UserOut, cnt int) []users2.UserOut {
	if len(available) == 1 {
		return available
	}

	shuffled := make([]users2.UserOut, len(available))
	copy(shuffled, available)

	u.randomizer.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:cnt]
}
