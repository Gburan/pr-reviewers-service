package pull_request_reassign

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
	"pr-reviewers-service/internal/usecase/pull_request_reassign"

	"github.com/go-playground/validator/v10"
)

type reassignPullRequestHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func New(usecase usecase, validator *validator.Validate) *reassignPullRequestHandler {
	return &reassignPullRequestHandler{
		usecase:   usecase,
		validator: validator,
	}
}

// @Summary Reassign pull request reviewer
// @Description Replace one reviewer with another from the same team
// @ID ReassignPullRequest
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param input body handler2.PostPullRequestReassignJSONRequestBody true "Reassignment data"
// @Success 200 {object} handler2.ReassignPullRequestResponse "Reviewer successfully reassigned"
// @Failure 400 {object} handler2.ErrorResponse "Invalid request data"
// @Failure 422 {object} handler2.ErrorResponse "Validation failed"
// @Failure 404 {object} handler2.ErrorResponse "Pull request, reviewer, author not found or no available reviewers"
// @Failure 409 {object} handler2.ErrorResponse "Pull request already merged"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /pullRequest/reassign [post]
func (h *reassignPullRequestHandler) ReassignPullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request handler2.PostPullRequestReassignJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "failed to decode request", err)
		return
	}

	if err := h.validator.Struct(request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusUnprocessableEntity, handler2.BADREQUEST, "validation failed", err)
		return
	}

	ctx = logging.WithLogPullRequestID(ctx, request.PullRequestId)

	result, err := h.usecase.Run(ctx, pull_request_reassign.In{
		PullRequestID: request.PullRequestId,
		OldUserId:     request.OldReviewerId,
	})
	if err != nil {
		h.handleUseCaseError(w, ctx, err)
		return
	}

	out := handler2.ReassignPullRequestResponse{
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
		ReplacedBy: result.ReplacedBy,
	}

	if err = json.NewEncoder(w).Encode(out); err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "failed to encode response", err)
		return
	}
}

func (h *reassignPullRequestHandler) handleUseCaseError(w http.ResponseWriter, ctx context.Context, err error) {
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
	case errors.Is(err, usecase2.ErrGetUser):
		errorMsg = "error occurred while getting user"
	case errors.Is(err, usecase2.ErrGetUsers):
		errorMsg = "error occurred while getting users"
	case errors.Is(err, usecase2.ErrRemoveReviewer):
		errorMsg = "error occurred while removing reviewer"
	case errors.Is(err, usecase2.ErrAssignReviewer):
		errorMsg = "error occurred while assigning reviewer"
	case errors.Is(err, usecase2.ErrPullRequestNotFound):
		errorMsg = "pull request not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	case errors.Is(err, usecase2.ErrReviewerNotFound):
		errorMsg = "reviewer not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	case errors.Is(err, usecase2.ErrNoAvailableReviewers):
		errorMsg = "no available reviewers"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOCANDIDATE
	case errors.Is(err, usecase2.ErrPullRequestAlreadyMerged):
		errorMsg = "pull request already merged"
		statusCode = http.StatusConflict
		errorResponseErrorCode = handler2.PRMERGED
	case errors.Is(err, usecase2.ErrAuthorPrNotFound):
		errorMsg = "author not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
