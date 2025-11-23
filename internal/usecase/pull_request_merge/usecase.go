package pull_request_merge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	pr_statuses2 "pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_statuses"
	"pr-reviewers-service/internal/usecase/contract/repository/pull_requests"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/google/uuid"
)

type usecase struct {
	repPullRequests pull_requests.RepositoryPullRequests
	repPRReviewers  pr_reviewers.RepositoryPrReviewers
	repPRStatuses   pr_statuses.RepositoryPrStatuses
	trm             trm.Manager
}

func NewUsecase(
	repPullRequests pull_requests.RepositoryPullRequests,
	repPRReviewers pr_reviewers.RepositoryPrReviewers,
	repPRStatuses pr_statuses.RepositoryPrStatuses,
	trm trm.Manager,
) *usecase {
	return &usecase{
		repPullRequests: repPullRequests,
		repPRReviewers:  repPRReviewers,
		repPRStatuses:   repPRStatuses,
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
	slog.DebugContext(ctx, "Get assigned reviewers")
	reviewers, err := u.repPRReviewers.GetPRReviewersByPRID(ctx, existingPR.ID)
	if err != nil && !errors.Is(err, repository.ErrPRReviewerNotFound) {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: pr_id %s", usecase2.ErrGetPRReviewers, existingPR.ID))
	}
	slog.DebugContext(ctx, "Check PR status", "current_status", currentStatus.Status, "merged_status", usecase2.MergedStatusValue)
	if currentStatus.Status == usecase2.MergedStatusValue {
		slog.DebugContext(ctx, "PR already merged, returning current state")
		return &Out{
			PullRequestID:   existingPR.ID,
			PullRequestName: existingPR.Name,
			AuthorID:        existingPR.AuthorID,
			Status:          usecase2.MergedStatusValue,
			AssignedReviewers: func() []uuid.UUID {
				if reviewers == nil {
					return nil
				}
				revs := make([]uuid.UUID, 0, len(*reviewers))
				for _, rev := range *reviewers {
					revs = append(revs, rev.ReviewerID)
				}
				return revs
			}(),
			CreatedAt: existingPR.CreatedAt,
			MergedAt:  existingPR.MergedAt,
		}, logging.WrapError(ctx, fmt.Errorf("%w: pr_id %s", usecase2.ErrPullRequestAlreadyMerged, existingPR.ID))
	}

	statusIn := pr_statuses2.PRStatusIn{
		ID:     currentStatus.ID,
		Status: usecase2.MergedStatusValue,
	}
	statusOut, err := u.repPRStatuses.UpdatePRStatusByID(ctx, statusIn)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrUpdatePrStatus, req.PullRequestID))
	}

	slog.DebugContext(ctx, "Update pull request status to MERGED")
	updatedPR, err := u.repPullRequests.MarkPullRequestMergedByID(ctx, existingPR.ID)
	if err != nil {
		return nil, logging.WrapError(ctx, fmt.Errorf("%w: %s", usecase2.ErrUpdatePrMergeTime, req.PullRequestID))
	}

	slog.DebugContext(ctx, "UseCase MergePullRequest success")
	return &Out{
		PullRequestID:   updatedPR.ID,
		PullRequestName: updatedPR.Name,
		AuthorID:        updatedPR.AuthorID,
		Status:          statusOut.Status,
		AssignedReviewers: func() []uuid.UUID {
			if reviewers == nil {
				return nil
			}
			revs := make([]uuid.UUID, 0, len(*reviewers))
			for _, rev := range *reviewers {
				revs = append(revs, rev.ReviewerID)
			}
			return revs
		}(),
		CreatedAt: updatedPR.CreatedAt,
		MergedAt:  updatedPR.MergedAt,
	}, nil
}
