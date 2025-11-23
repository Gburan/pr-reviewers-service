package pull_request_create

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	pr_reviewers2 "pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	pr_statuses2 "pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	pull_requests2 "pr-reviewers-service/internal/infrastructure/repository/pull_requests"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	"pr-reviewers-service/internal/logging"
	"pr-reviewers-service/internal/metrics"
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
	slog.DebugContext(ctx, "Check if PR already exists")
	existingPR, err := u.repPullRequests.GetPullRequestByID(ctx, req.PullRequestID)
	if err != nil && !errors.Is(err, repository.ErrPullRequestNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetPullRequest, req.PullRequestID))
	}
	if existingPR != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrPullRequestExists, req.PullRequestID))
	}

	slog.DebugContext(ctx, "Get author", "author_id", req.AuthorID)
	author, err := u.repUsers.GetUserByID(ctx, req.AuthorID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrAuthorPrNotFound, req.AuthorID))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetUser, req.AuthorID))
	}

	slog.DebugContext(ctx, "Get active team members", "team_id", author.TeamID)
	teamMembers, err := u.repUsers.GetActiveUsersByTeamID(ctx, author.TeamID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: team_id %s", usecase2.ErrGetUsers, author.TeamID))
	}

	statusIn := pr_statuses2.PRStatusIn{
		Status: usecase2.OpenStatusValue,
	}
	prStatusOut, err := u.repPRStatuses.SavePRStatus(ctx, statusIn)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrSetPRStatus))
	}

	slog.DebugContext(ctx, "Create pull request")
	prIn := pull_requests2.PullRequestIn{
		ID:       req.PullRequestID,
		Name:     req.PullRequestName,
		AuthorID: req.AuthorID,
		StatusID: prStatusOut.ID,
	}
	createdPR, err := u.repPullRequests.SavePullRequest(ctx, prIn)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrSavePullRequest, req.PullRequestID))
	}

	availableReviewers := u.getAvailableReviewersFromTeam(teamMembers, req.AuthorID)
	selectedReviewers := u.selectRandomReviewers(availableReviewers, u.maxCntReviewers)
	slog.DebugContext(ctx, "Assign reviewers", "count", len(selectedReviewers))
	var assignedReviewers []uuid.UUID
	for _, reviewer := range selectedReviewers {
		reviewerIn := pr_reviewers2.PrReviewerIn{
			PrID:       createdPR.ID,
			ReviewerID: reviewer.ID,
		}
		_, err = u.repPRReviewers.SavePRReviewer(ctx, reviewerIn)
		if err != nil {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: reviewer %s", usecase2.ErrAssignReviewer, reviewer.ID))
		}
		assignedReviewers = append(assignedReviewers, reviewer.ID)
	}

	metrics.IncCreatedPRs()
	slog.DebugContext(ctx, "UseCase CreatePullRequest success", "reviewers_count", len(assignedReviewers))
	return &Out{
		PullRequestID:     createdPR.ID,
		PullRequestName:   createdPR.Name,
		AuthorID:          createdPR.AuthorID,
		Status:            prStatusOut.Status,
		AssignedReviewers: assignedReviewers,
		CreatedAt:         createdPR.CreatedAt,
		MergedAt:          createdPR.MergedAt,
	}, nil
}

func (u *usecase) getAvailableReviewersFromTeam(teamMembers *[]users2.UserOut, authorID uuid.UUID) []users2.UserOut {
	if teamMembers == nil {
		return []users2.UserOut{}
	}

	var available []users2.UserOut
	for _, member := range *teamMembers {
		if member.ID != authorID && member.IsActive {
			available = append(available, member)
		}
	}
	return available
}

func (u *usecase) selectRandomReviewers(available []users2.UserOut, maxCount int) []users2.UserOut {
	if len(available) <= maxCount {
		return available
	}

	shuffled := make([]users2.UserOut, len(available))
	copy(shuffled, available)

	u.randomizer.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:maxCount]
}
