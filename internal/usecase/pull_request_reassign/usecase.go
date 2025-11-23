package pull_request_reassign

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	pr_reviewers2 "pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	pr_statuses2 "pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/contract/randomizer"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_statuses"
	"pr-reviewers-service/internal/usecase/contract/repository/pull_requests"
	"pr-reviewers-service/internal/usecase/contract/repository/users"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/google/uuid"
)

type usecase struct {
	repUsers        users.RepositoryUsers
	repPullRequests pull_requests.RepositoryPullRequests
	repPRReviewers  pr_reviewers.RepositoryPrReviewers
	repPRStatuses   pr_statuses.RepositoryPrStatuses
	randomizer      randomizer.Randomizer
	maxCntReviewers int
	trm             trm.Manager
}

func NewUsecase(
	repUsers users.RepositoryUsers,
	repPullRequests pull_requests.RepositoryPullRequests,
	repPRReviewers pr_reviewers.RepositoryPrReviewers,
	repPRStatuses pr_statuses.RepositoryPrStatuses,
	randomizer randomizer.Randomizer,
	maxCntReviewers int,
	trm trm.Manager,
) *usecase {
	return &usecase{
		repUsers:        repUsers,
		repPullRequests: repPullRequests,
		repPRReviewers:  repPRReviewers,
		repPRStatuses:   repPRStatuses,
		randomizer:      randomizer,
		maxCntReviewers: maxCntReviewers,
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
	slog.DebugContext(ctx, "Get pull request", "pull_request_id", req.PullRequestID)
	existingPR, err := u.repPullRequests.GetPullRequestByID(ctx, req.PullRequestID)
	if err != nil {
		if errors.Is(err, repository.ErrPullRequestNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrPullRequestNotFound, req.PullRequestID))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetPullRequest, req.PullRequestID))
	}

	currentStatus, err := u.repPRStatuses.GetPRStatusByID(ctx, pr_statuses2.PRStatusIn{ID: existingPR.StatusID})
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: status_id %s", usecase2.ErrGetPRStatus, existingPR.StatusID))
	}
	if currentStatus.Status == usecase2.MergedStatusValue {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: pr_id %s", usecase2.ErrPullRequestAlreadyMerged, existingPR.ID))
	}

	slog.DebugContext(ctx, "Get current reviewers")
	currentReviewers, err := u.repPRReviewers.GetPRReviewersByPRID(ctx, existingPR.ID)
	if err != nil && !errors.Is(err, repository.ErrPRReviewerNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: pr_id %s", usecase2.ErrGetPRReviewers, existingPR.ID))
	}
	slog.DebugContext(ctx, "Check if old reviewer exists in current reviewers",
		"old_reviewer_id", req.OldUserId,
		"current_reviewer_count", len(*currentReviewers),
		"current_reviewer_ids", func() []string {
			ids := make([]string, 0, len(*currentReviewers))
			for _, reviewer := range *currentReviewers {
				ids = append(ids, reviewer.ID.String())
			}
			return ids
		}())
	oldReviewerExists := false
	for _, reviewer := range *currentReviewers {
		if reviewer.ReviewerID == req.OldUserId {
			oldReviewerExists = true
			break
		}
	}
	if !oldReviewerExists {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: reviewer_id %s", usecase2.ErrReviewerNotFound, req.OldUserId))
	}

	slog.DebugContext(ctx, "Get author", "author_id", existingPR.AuthorID)
	author, err := u.repUsers.GetUserByID(ctx, existingPR.AuthorID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrAuthorPrNotFound, existingPR.AuthorID))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetUser, existingPR.AuthorID))
	}
	slog.DebugContext(ctx, "Get active team members", "team_id", author.TeamID)
	teamMembers, err := u.repUsers.GetActiveUsersByTeamID(ctx, author.TeamID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: team_id %s", usecase2.ErrGetUsers, author.TeamID))
	}
	availableReviewers := u.getAvailableReviewersFromTeam(*teamMembers, existingPR.AuthorID, *currentReviewers)
	if len(availableReviewers) == 0 {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: team_id %s", usecase2.ErrNoAvailableReviewers, author.TeamID))
	}
	newReviewer := u.selectRandomReviewer(availableReviewers)

	slog.DebugContext(ctx, "Remove old reviewer", "old_reviewer_id", req.OldUserId)
	err = u.repPRReviewers.DeletePRReviewerByPRAndReviewer(ctx, existingPR.ID, req.OldUserId)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: reviewer_id %s", usecase2.ErrRemoveReviewer, req.OldUserId))
	}
	slog.DebugContext(ctx, "Assign new reviewer", "new_reviewer_id", newReviewer.ID)
	reviewerIn := pr_reviewers2.PrReviewerIn{
		PrID:       existingPR.ID,
		ReviewerID: newReviewer.ID,
	}
	_, err = u.repPRReviewers.SavePRReviewer(ctx, reviewerIn)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: reviewer %s", usecase2.ErrAssignReviewer, newReviewer.ID))
	}

	slog.DebugContext(ctx, "Get updated reviewers list")
	updatedReviewers, err := u.repPRReviewers.GetPRReviewersByPRID(ctx, existingPR.ID)
	if err != nil && !errors.Is(err, repository.ErrPRReviewerNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: pr_id %s", usecase2.ErrGetPRReviewers, existingPR.ID))
	}
	var assignedReviewers []uuid.UUID
	if updatedReviewers != nil {
		for _, reviewer := range *updatedReviewers {
			assignedReviewers = append(assignedReviewers, reviewer.ReviewerID)
		}
	}

	slog.DebugContext(ctx, "UseCase ReassignPullRequest success", "replaced_by", newReviewer.ID)
	return &Out{
		PullRequestID:     existingPR.ID,
		PullRequestName:   existingPR.Name,
		AuthorID:          existingPR.AuthorID,
		Status:            currentStatus.Status,
		AssignedReviewers: assignedReviewers,
		CreatedAt:         existingPR.CreatedAt,
		MergedAt:          existingPR.MergedAt,
		ReplacedBy:        newReviewer.ID,
	}, nil
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

func (u *usecase) selectRandomReviewer(available []users2.UserOut) users2.UserOut {
	if len(available) == 1 {
		return available[0]
	}

	shuffled := make([]users2.UserOut, len(available))
	copy(shuffled, available)

	u.randomizer.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[0]
}
