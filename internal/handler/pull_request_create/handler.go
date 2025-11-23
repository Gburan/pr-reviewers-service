package pull_request_create

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
	"pr-reviewers-service/internal/usecase/pull_request_create"

	"github.com/go-playground/validator/v10"
)

type createPullRequestHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func New(usecase usecase, validator *validator.Validate) *createPullRequestHandler {
	return &createPullRequestHandler{
		usecase:   usecase,
		validator: validator,
	}
}

// @Summary Create pull request
// @Description Create PR and automatically assign up to 2 reviewers from author's team
// @ID CreatePullRequest
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param input body handler2.PostPullRequestCreateJSONRequestBody true "Pull request data"
// @Success 201 {object} handler2.CreatePullRequestResponse "PR successfully created"
// @Failure 400 {object} handler2.ErrorResponse "Bad request (PR already exists or invalid data)"
// @Failure 422 {object} handler2.ErrorResponse "Validation failed"
// @Failure 404 {object} handler2.ErrorResponse "Author not found"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /pullRequest/create [post]
func (h *createPullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request handler2.PostPullRequestCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "failed to decode request", err)
		return
	}

	if err := h.validator.Struct(request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusUnprocessableEntity, handler2.BADREQUEST, "validation failed", err)
		return
	}

	ctx = logging.WithLogAuthorID(ctx, request.AuthorId)
	ctx = logging.WithLogPullRequestID(ctx, request.PullRequestId)

	result, err := h.usecase.Run(ctx, pull_request_create.In{
		PullRequestID:   request.PullRequestId,
		PullRequestName: request.PullRequestName,
		AuthorID:        request.AuthorId,
	})
	if err != nil {
		h.handleUseCaseError(w, ctx, err)
		return
	}

	out := handler2.CreatePullRequestResponse{
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
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(out); err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "failed to encode response", err)
		return
	}
}

func (h *createPullRequestHandler) handleUseCaseError(w http.ResponseWriter, ctx context.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorResponseErrorCode := handler2.UNKNOWN
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase2.ErrGetPullRequest):
		errorMsg = "error occurred while checking pull request existence"
	case errors.Is(err, usecase2.ErrGetUser):
		errorMsg = "error occurred while getting author information"
	case errors.Is(err, usecase2.ErrGetUsers):
		errorMsg = "error occurred while getting team members"
	case errors.Is(err, usecase2.ErrSetPRStatus):
		errorMsg = "error occurred while setting PR status"
	case errors.Is(err, usecase2.ErrSavePullRequest):
		errorMsg = "error occurred while saving pull request in db"
	case errors.Is(err, usecase2.ErrAssignReviewer):
		errorMsg = "error occurred while assigning reviewers"
	case errors.Is(err, usecase2.ErrAuthorPrNotFound):
		errorMsg = "author not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOCANDIDATE
	case errors.Is(err, usecase2.ErrPullRequestExists):
		errorMsg = "pull request already exists"
		statusCode = http.StatusBadRequest
		errorResponseErrorCode = handler2.PREXISTS
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
