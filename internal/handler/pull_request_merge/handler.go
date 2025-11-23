package pull_request_merge

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/pull_request_merge"

	"github.com/go-playground/validator/v10"
)

type mergePullRequestHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func New(usecase usecase, validator *validator.Validate) *mergePullRequestHandler {
	return &mergePullRequestHandler{
		usecase:   usecase,
		validator: validator,
	}
}

// @Summary Merge pull request
// @Description Merge an existing pull request
// @ID MergePullRequest
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param input body handler2.PostPullRequestMergeJSONRequestBody true "Pull request merge data"
// @Success 200 {object} handler2.MergePullRequestResponse "PR successfully merged or already merged"
// @Failure 400 {object} handler2.ErrorResponse "Invalid request data"
// @Failure 422 {object} handler2.ErrorResponse "Validation failed"
// @Failure 404 {object} handler2.ErrorResponse "Pull request not found"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /pullRequest/merge [post]
func (h *mergePullRequestHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request handler2.PostPullRequestMergeJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "failed to decode request", err)
		return
	}

	if err := h.validator.Struct(request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusUnprocessableEntity, handler2.BADREQUEST, "validation failed", err)
		return
	}

	ctx = logging.WithLogPullRequestID(ctx, request.PullRequestId)

	result, err := h.usecase.Run(ctx, pull_request_merge.In{
		PullRequestID: request.PullRequestId,
	})
	if err != nil && !errors.Is(err, usecase2.ErrPullRequestAlreadyMerged) {
		h.handleUseCaseError(w, ctx, err)
		return
	}

	out := handler2.MergePullRequestResponse{
		Pr: handler2.PullRequest{
			PullRequestId:     result.PullRequestID,
			PullRequestName:   result.PullRequestName,
			AuthorId:          result.AuthorID,
			Status:            handler2.PullRequestStatus(result.Status),
			AssignedReviewers: result.AssignedReviewers,
			CreatedAt:         &result.CreatedAt,
			MergedAt: func() *time.Time {
				if result.MergedAt.IsZero() {
					return nil
				}
				return &result.MergedAt
			}(),
		},
	}

	if err = json.NewEncoder(w).Encode(out); err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "failed to encode response", err)
		return
	}
}

func (h *mergePullRequestHandler) handleUseCaseError(w http.ResponseWriter, ctx context.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorResponseErrorCode := handler2.UNKNOWN
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase2.ErrGetPullRequest):
		errorMsg = "error occurred while getting pull request"
	case errors.Is(err, usecase2.ErrGetPRStatus):
		errorMsg = "error occurred while getting pr status"
	case errors.Is(err, usecase2.ErrGetPRReviewers):
		errorMsg = "error occurred while getting pr reviewers"
	case errors.Is(err, usecase2.ErrSetPRStatus):
		errorMsg = "error occurred while setting pr status"
	case errors.Is(err, usecase2.ErrUpdatePrMergeTime):
		errorMsg = "error occurred while updating pr merge time"
	case errors.Is(err, usecase2.ErrPullRequestNotFound):
		errorMsg = "pull request not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
