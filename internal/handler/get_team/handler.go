package get_team

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/logging"
	usecase2 "pr-reviewers-service/internal/usecase"
	"pr-reviewers-service/internal/usecase/get_team"
)

type getTeamHandler struct {
	usecase usecase
}

func New(usecase usecase) *getTeamHandler {
	return &getTeamHandler{
		usecase: usecase,
	}
}

// @Summary Get team information
// @Description Get team details with members by team name
// @ID GetTeam
// @Tags Teams
// @Accept json
// @Produce json
// @Param team_name query string true "Team name"
// @Success 200 {object} handler2.Team "Team found successfully"
// @Failure 400 {object} handler2.ErrorResponse "Invalid team_name parameter"
// @Failure 404 {object} handler2.ErrorResponse "Team not found"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /team/get [get]
func (h *getTeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request handler2.GetTeamGetParams
	queryParams := r.URL.Query()
	teamName := queryParams.Get("team_name")
	if strings.TrimSpace(teamName) == "" {
		handler.RespondWithError(w, ctx, http.StatusBadRequest, handler2.BADREQUEST, "team_name cannot be empty", nil)
		return
	}
	request.TeamName = handler2.TeamNameQuery(teamName)

	ctx = logging.WithLogTeamName(ctx, request.TeamName)

	result, err := h.usecase.Run(ctx, get_team.In{
		TeamName: request.TeamName,
	})
	if err != nil {
		handleUseCaseError(w, ctx, err)
		return
	}

	out := handler2.Team{
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
	if err = json.NewEncoder(w).Encode(out); err != nil {
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
	case errors.Is(err, usecase2.ErrGetUsers):
		errorMsg = "error occurred while checking if users exist with such ids"
	case errors.Is(err, usecase2.ErrTeamNotFound):
		errorMsg = "there is no team looking for"
		statusCode = http.StatusNotFound
		errorResponseErrorCode = handler2.NOTFOUND
	}

	handler.RespondWithError(w, ctx, statusCode, errorResponseErrorCode, errorMsg, err)
}
