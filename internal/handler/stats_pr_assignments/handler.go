package stats_pr_assignments

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/stats_pr_assignments"
)

type reviewersStatsHandler struct {
	usecase usecase
}

func New(usecase usecase) *reviewersStatsHandler {
	return &reviewersStatsHandler{
		usecase: usecase,
	}
}

// @Summary Get reviewers assignment statistics
// @Description Get assignment count statistics for all reviewers
// @ID GetReviewersStats
// @Tags Statistics
// @Produce json
// @Success 200 {object} handler2.ReviewersStatsResponse "Statistics successfully retrieved"
// @Failure 404 {object} handler2.ErrorResponse "Reviewers data not found"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /stats/reviewers [get]
func (h *reviewersStatsHandler) GetReviewersStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	result, err := h.usecase.Run(ctx, stats_pr_assignments.In{})
	if err != nil {
		h.handleUseCaseError(w, ctx, err)
		return
	}

	reviewers := make([]handler2.ReviewerAssignmentCount, 0, len(result.Reviewers))
	for _, reviewer := range result.Reviewers {
		reviewers = append(reviewers, handler2.ReviewerAssignmentCount{
			ReviewerId:      reviewer.ReviewerID,
			AssignmentCount: reviewer.AssignmentCount,
		})
	}

	out := handler2.ReviewersStatsResponse{
		Reviewers: reviewers,
	}
	if err = json.NewEncoder(w).Encode(out); err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "failed to encode response", err)
		return
	}
}

func (h *reviewersStatsHandler) handleUseCaseError(w http.ResponseWriter, ctx context.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorResponseErrorCode := handler2.UNKNOWN
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase2.ErrGetPRReviewers):
		errorMsg = "error occurred while getting PR reviewers"
	case errors.Is(err, usecase2.ErrPRsReviewersNotFound):
		errorMsg = "PR reviewers not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
