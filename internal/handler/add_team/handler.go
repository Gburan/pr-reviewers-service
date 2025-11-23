package add_team

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/add_team"

	"github.com/go-playground/validator/v10"
)

type addTeamHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func New(usecase usecase, validator *validator.Validate) *addTeamHandler {
	return &addTeamHandler{
		usecase:   usecase,
		validator: validator,
	}
}

// @Summary Create team with members
// @Description Create a new team with members (creates/updates users)
// @ID AddTeam
// @Tags Teams
// @Accept json
// @Produce json
// @Param input body handler2.PostTeamAddJSONRequestBody true "Team data with members"
// @Success 201 {object} handler2.Team "Team successfully created"
// @Success 304 "No changes - team exists and no users were changed or added"
// @Failure 400 {object} handler2.ErrorResponse "Validation failed or duplicate users"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /team/add [post]
func (h *addTeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request handler2.PostTeamAddJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "failed to decode request", err)
		return
	}

	if err := h.validator.Struct(request); err != nil {
		handler.RespondWithError(w, ctx, http.StatusUnprocessableEntity, handler2.BADREQUEST, "validation failed", err)
		return
	}

	ctx = logging.WithLogTeamName(ctx, request.TeamName)
	ctx = logging.WithLogTeamMembersCount(ctx, len(request.Members))

	result, err := h.usecase.Run(ctx, add_team.In{
		TeamName: request.TeamName,
		Members: func() []add_team.TeamMembers {
			members := make([]add_team.TeamMembers, 0, len(request.Members))
			for _, member := range request.Members {
				members = append(members, add_team.TeamMembers{
					IsActive: member.IsActive,
					UserID:   member.UserId,
					Username: member.Username,
				})
			}
			return members
		}(),
	})
	if err != nil {
		handleUseCaseError(w, ctx, err)
		return
	}

	team := handler2.Team{
		TeamName: result.TeamName,
		Members: func() []handler2.TeamMember {
			members := make([]handler2.TeamMember, 0, len(result.Members))
			for _, member := range result.Members {
				members = append(members, handler2.TeamMember{
					IsActive: member.IsActive,
					UserId:   member.UserID,
					Username: member.Username,
				})
			}
			return members
		}(),
	}

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(team); err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "failed to encode response", err)
		return
	}
}

func handleUseCaseError(w http.ResponseWriter, ctx context.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorResponseErrorCode := handler2.UNKNOWN
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase2.ErrGetTeam):
		errorMsg = "error occurred while getting team with such id"
	case errors.Is(err, usecase2.ErrSaveTeam):
		errorMsg = "error occurred while saving team in db"
	case errors.Is(err, usecase2.ErrGetUsers):
		errorMsg = "error occurred while checking if users exist with such ids"
	case errors.Is(err, usecase2.ErrSaveUsersBatch):
		errorMsg = "error occurred while saving new users in db"
	case errors.Is(err, usecase2.ErrUpdateUsersBatch):
		errorMsg = "error occurred while updating existing users in db"
	case errors.Is(err, usecase2.ErrNoUsersWereUpdatedAddedTeam):
		errorMsg = "team exists and no users were changed or added"
		errorResponseErrorCode = handler2.TEAMEXISTS
		statusCode = http.StatusNotModified
	case errors.Is(err, usecase2.ErrDuplicateUsers):
		errorMsg = "dont use same user ids"
		errorResponseErrorCode = handler2.BADREQUEST
		statusCode = http.StatusBadRequest
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
