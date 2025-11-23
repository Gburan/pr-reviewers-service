package team_deactivate_users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/team_deactivate_users"

	"github.com/go-playground/validator/v10"
)

type deactivateTeamUsersHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func New(usecase usecase, validator *validator.Validate) *deactivateTeamUsersHandler {
	return &deactivateTeamUsersHandler{
		usecase:   usecase,
		validator: validator,
	}
}

// @Summary Deactivate team users
// @Description Deactivate multiple users in a team and handle PR reassignments
// @ID DeactivateTeamUsers
// @Tags Teams
// @Accept json
// @Produce json
// @Param input body handler2.PatchTeamDeactivateUsersJSONRequestBody true "Team deactivation data"
// @Success 200 {object} handler2.DeactivateTeamUsersResponse "Users successfully deactivated"
// @Failure 400 {object} handler2.ErrorResponse "Invalid request data"
// @Failure 422 {object} handler2.ErrorResponse "Validation failed"
// @Failure 404 {object} handler2.ErrorResponse "Team, users, or pull requests not found"
// @Failure 409 {object} handler2.ErrorResponse "User does not belong to the specified team"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /team/deactivateUsers [patch]
func (h *deactivateTeamUsersHandler) DeactivateTeamUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request handler2.PatchTeamDeactivateUsersJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "failed to decode request", err)
		return
	}

	if err := h.validator.Struct(request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusUnprocessableEntity, handler2.BADREQUEST, "validation failed", err)
		return
	}

	ctx = logging.WithLogTeamName(ctx, request.TeamName)

	result, err := h.usecase.Run(ctx, team_deactivate_users.In{
		TeamName: request.TeamName,
		UserIDs:  request.UserIds,
	})
	if err != nil {
		h.handleUseCaseError(w, ctx, err)
		return
	}

	out := handler2.DeactivateTeamUsersResponse{
		Team: handler2.Team{
			TeamName: result.Team.TeamName,
			Members: func() []handler2.TeamMember {
				members := make([]handler2.TeamMember, 0, len(result.Team.Members))
				for _, member := range result.Team.Members {
					members = append(members, handler2.TeamMember{
						UserId:   member.UserID,
						Username: member.Username,
						IsActive: member.IsActive,
					})
				}
				return members
			}(),
		},
		AffectedPullRequests: func() []handler2.PullRequestShort {
			prs := make([]handler2.PullRequestShort, 0, len(result.AffectedPullRequests))
			for _, pr := range result.AffectedPullRequests {
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

func (h *deactivateTeamUsersHandler) handleUseCaseError(w http.ResponseWriter, ctx context.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorResponseErrorCode := handler2.UNKNOWN
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase2.ErrGetTeam):
		errorMsg = "error occurred while getting team"
	case errors.Is(err, usecase2.ErrGetUsers):
		errorMsg = "error occurred while getting users"
	case errors.Is(err, usecase2.ErrUpdateUser):
		errorMsg = "error occurred while updating user"
	case errors.Is(err, usecase2.ErrGetPRReviewers):
		errorMsg = "error occurred while getting pr reviewers"
	case errors.Is(err, usecase2.ErrGetPullRequest):
		errorMsg = "error occurred while getting pull request"
	case errors.Is(err, usecase2.ErrGetPRStatus):
		errorMsg = "error occurred while getting pr status"
	case errors.Is(err, usecase2.ErrRemoveReviewer):
		errorMsg = "error occurred while removing reviewer"
	case errors.Is(err, usecase2.ErrAssignReviewer):
		errorMsg = "error occurred while assigning reviewer"
	case errors.Is(err, usecase2.ErrUsersByIDsNotFound):
		errorMsg = "users not found by provided IDs"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	case errors.Is(err, usecase2.ErrTeamNotFound):
		errorMsg = "team not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	case errors.Is(err, usecase2.ErrUserNotBelongsToTeam):
		errorMsg = "user does not belong to the specified team"
		statusCode = http.StatusConflict
		errorResponseErrorCode = handler2.BADREQUEST
	case errors.Is(err, usecase2.ErrUserNotFound):
		errorMsg = "user not found"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	case errors.Is(err, usecase2.ErrNoPRsToAffect):
		errorMsg = "no pull requests to affect"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	case errors.Is(err, usecase2.ErrNoUsersAssignedToPRs):
		errorMsg = "no users assigned to pull requests"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
