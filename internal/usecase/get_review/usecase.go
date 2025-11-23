package get_review

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_statuses"
	"pr-reviewers-service/internal/usecase/contract/repository/pull_requests"
	"pr-reviewers-service/internal/usecase/contract/repository/users"

	"github.com/google/uuid"
)

type usecase struct {
	repUsers        users.RepositoryUsers
	repPullRequests pull_requests.RepositoryPullRequests
	repPRReviewers  pr_reviewers.RepositoryPrReviewers
	repPRStatuses   pr_statuses.RepositoryPrStatuses
}

func NewUsecase(
	repUsers users.RepositoryUsers,
	repPullRequests pull_requests.RepositoryPullRequests,
	repPRReviewers pr_reviewers.RepositoryPrReviewers,
	repPRStatuses pr_statuses.RepositoryPrStatuses,
) *usecase {
	return &usecase{
		repUsers:        repUsers,
		repPullRequests: repPullRequests,
		repPRReviewers:  repPRReviewers,
		repPRStatuses:   repPRStatuses,
	}
}

func (u *usecase) Run(ctx context.Context, req In) (*Out, error) {
	slog.DebugContext(ctx, "Get user", "author_id", req.UserID)
	_, err := u.repUsers.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrUserNotFound, req.UserID))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrGetUser, req.UserID))
	}

	slog.DebugContext(ctx, "Get PR reviewers by user ID", "user_id", req.UserID)
	prReviewers, err := u.repPRReviewers.GetPRReviewersByReviewerID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrPRReviewerNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w: user_id %s", usecase2.ErrNoActiveReviewers, req.UserID))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: user_id %s", usecase2.ErrGetPRReviewers, req.UserID))
	}
	slog.DebugContext(ctx, "Found PR reviewers", "count", len(*prReviewers))

	prIDs := make([]uuid.UUID, 0, len(*prReviewers))
	for _, prReviewer := range *prReviewers {
		prIDs = append(prIDs, prReviewer.PRID)
	}
	slog.DebugContext(ctx, "Get pull requests by IDs", "pr_ids_count", len(prIDs))
	pullRequestsList, err := u.repPullRequests.GetPullRequestsByPrIDs(ctx, prIDs)
	if err != nil {
		if errors.Is(err, repository.ErrPullRequestNotFound) {
			slog.DebugContext(ctx, "No pull requests found")
			return &Out{
				UserID:       req.UserID,
				PullRequests: []PullRequestShort{},
			}, nil
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrGetPullRequest))
	}
	slog.DebugContext(ctx, "Found pull requests", "count", len(*pullRequestsList))

	statusIDs := make([]uuid.UUID, 0, len(*pullRequestsList))
	for _, pr := range *pullRequestsList {
		statusIDs = append(statusIDs, pr.StatusID)
	}
	slog.DebugContext(ctx, "Get PR statuses by IDs", "status_ids_count", len(statusIDs))
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

	pullRequests := make([]PullRequestShort, 0, len(*pullRequestsList))
	for _, pr := range *pullRequestsList {
		status, exists := statusMap[pr.StatusID]
		if !exists {
			slog.DebugContext(ctx, "Status not found for PR, skipping", "pr_id", pr.ID, "status_id", pr.StatusID)
			continue
		}

		pullRequests = append(pullRequests, PullRequestShort{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.AuthorID,
			Status:          status,
		})
	}

	slog.DebugContext(ctx, "UseCase GetUserReviewPRs success", "pull_requests_count", len(pullRequests))
	return &Out{
		UserID:       req.UserID,
		PullRequests: pullRequests,
	}, nil
}
