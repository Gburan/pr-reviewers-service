package stats_pr_assignments

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"pr-reviewers-service/internal/infrastructure/repository"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers"

	"github.com/google/uuid"
)

type usecase struct {
	repPRReviewers pr_reviewers.RepositoryPrReviewers
}

func NewUsecase(repPRReviewers pr_reviewers.RepositoryPrReviewers) *usecase {
	return &usecase{
		repPRReviewers: repPRReviewers,
	}
}

func (u *usecase) Run(ctx context.Context, _ In) (*Out, error) {
	slog.DebugContext(ctx, "Call GetAllPRReviewers")
	allReviewers, err := u.repPRReviewers.GetAllPRReviewers(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrPRReviewerNotFound) {
			return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrPRsReviewersNotFound))
		}
		return nil, logging.WrapError(ctx, fmt.Errorf("%w", usecase2.ErrGetPRReviewers))
	}

	slog.DebugContext(ctx, "Calculate reviewers statistics", "total_assignments", len(*allReviewers))
	statsMap := make(map[uuid.UUID]int)
	for _, assignment := range *allReviewers {
		statsMap[assignment.ReviewerID]++
	}

	reviewers := make([]ReviewerStats, 0, len(statsMap))
	for reviewerID, count := range statsMap {
		reviewers = append(reviewers, ReviewerStats{
			ReviewerID:      reviewerID,
			AssignmentCount: count,
		})
	}

	slog.DebugContext(ctx, "UseCase Statistics success", "unique_reviewers", len(reviewers))
	return &Out{
		Reviewers: reviewers,
	}, nil
}
