package set_is_active

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/set_is_active"

	"github.com/go-playground/validator/v10"
)

type setIsActiveHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func New(usecase usecase, validator *validator.Validate) *setIsActiveHandler {
	return &setIsActiveHandler{
		usecase:   usecase,
		validator: validator,
	}
}

// @Summary Set user active status
// @Description Activate or deactivate a user
// @ID SetIsActive
// @Tags Users
// @Accept json
// @Produce json
// @Param input body handler2.PostUsersSetIsActiveJSONRequestBody true "User status data"
// @Success 200 {object} handler2.SetUserActiveStatusResponse "User status successfully updated"
// @Success 304 "User already has the target status (no changes)"
// @Failure 400 {object} handler2.ErrorResponse "Invalid request data"
// @Failure 422 {object} handler2.ErrorResponse "Validation failed"
// @Failure 404 {object} handler2.ErrorResponse "User not found"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /users/setIsActive [post]
func (h *setIsActiveHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request handler2.PostUsersSetIsActiveJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "failed to decode request", err)
		return
	}

	if err := h.validator.Struct(request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusUnprocessableEntity, handler2.BADREQUEST, "validation failed", err)
		return
	}

	ctx = logging.WithLogUserId(ctx, request.UserId)

	result, err := h.usecase.Run(ctx, set_is_active.In{
		UserID:   request.UserId,
		IsActive: request.IsActive,
	})
	if err != nil {
		h.handleUseCaseError(w, ctx, err)
		return
	}

	out := handler2.SetUserActiveStatusResponse{
		User: handler2.User{
			UserId:   result.UserId,
			Username: result.Username,
			TeamName: result.TeamName,
			IsActive: result.IsActive,
		},
	}

	if err = json.NewEncoder(w).Encode(out); err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "failed to encode response", err)
		return
	}
}

func (h *setIsActiveHandler) handleUseCaseError(w http.ResponseWriter, ctx context.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorResponseErrorCode := handler2.UNKNOWN
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase2.ErrUserNotFound):
		errorMsg = "user not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	case errors.Is(err, usecase2.ErrUpdateUser):
		errorMsg = "error occurred while updating user in db"
	case errors.Is(err, usecase2.ErrGetUser):
		errorMsg = "error occurred while getting user from db"
	case errors.Is(err, usecase2.ErrGetTeam):
		errorMsg = "error occurred while getting users team"
	case errors.Is(err, usecase2.ErrUserDontNeedChange):
		errorMsg = "user already have same status as you trying to assign"
		statusCode = http.StatusNotModified
		errorResponseErrorCode = handler2.NOTASSIGNED
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
