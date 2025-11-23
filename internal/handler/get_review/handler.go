package get_review

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/get_review"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type getUserReviewPRsHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func New(usecase usecase, validator *validator.Validate) *getUserReviewPRsHandler {
	return &getUserReviewPRsHandler{
		usecase:   usecase,
		validator: validator,
	}
}

// @Summary Get users pull requests for review
// @Description Get all pull requests assigned to user for review
// @ID GetUserReviewPRs
// @Tags Reviews
// @Accept json
// @Produce json
// @Param user_id query string true "User ID" format(uuid)
// @Success 200 {object} handler2.GetUserReviewPRsResponse "Successfully retrieved pull requests"
// @Failure 400 {object} handler2.ErrorResponse "Missing or invalid user_id"
// @Failure 404 {object} handler2.ErrorResponse "User not found or no active reviewers"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /user/review [get]
func (h *getUserReviewPRsHandler) GetUserReviewPRs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "user_id is required", nil)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "invalid user_id format", err)
		return
	}

	ctx = logging.WithLogUserId(ctx, userID)

	result, err := h.usecase.Run(ctx, get_review.In{
		UserID: userID,
	})
	if err != nil {
		h.handleUseCaseError(w, ctx, err)
		return
	}

	out := handler2.GetUserReviewPRsResponse{
		UserId: userID,
		PullRequests: func() []handler2.PullRequestShort {
			prs := make([]handler2.PullRequestShort, 0, len(result.PullRequests))
			for _, pr := range result.PullRequests {
				prs = append(prs, handler2.PullRequestShort{
					PullRequestId:   pr.PullRequestID,
					PullRequestName: pr.PullRequestName,
					AuthorId:        pr.AuthorID,
					Status:          handler2.PullRequestShortStatus(pr.Status),
				})
			}
			return prs
		}(),
	}

	if err = json.NewEncoder(w).Encode(out); err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "failed to encode response", err)
		return
	}
}

func (h *getUserReviewPRsHandler) handleUseCaseError(w http.ResponseWriter, ctx context.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorResponseErrorCode := handler2.UNKNOWN
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase2.ErrGetPRReviewers):
		errorMsg = "error occurred while getting pr ids where user reviewer"
	case errors.Is(err, usecase2.ErrGetPullRequest):
		errorMsg = "error occurred while getting pull requests"
	case errors.Is(err, usecase2.ErrGetUser):
		errorMsg = "error occurred while getting user from db"
	case errors.Is(err, usecase2.ErrGetPRStatus):
		errorMsg = "error occurred while getting pr statuses"
	case errors.Is(err, usecase2.ErrUserNotFound):
		errorMsg = "user not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	case errors.Is(err, usecase2.ErrNoActiveReviewers):
		errorMsg = "user is not assigned as reviewer to any pull requests"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTASSIGNED
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
