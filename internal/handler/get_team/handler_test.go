package get_team_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "pr-reviewers-service/internal/generated/api/v1/handler"
	get_team_handler "pr-reviewers-service/internal/handler/get_team"
	mock_get_team "pr-reviewers-service/internal/handler/get_team/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecase "pr-reviewers-service/internal/usecase/get_team"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mock_get_team.NewMockusecase(ctrl)
	h := get_team_handler.New(mockUC)

	teamName := "backend"
	userID := uuid.New()

	ucIn := usecase.In{TeamName: teamName}

	ucOut := usecase.Out{
		TeamName: teamName,
		Members: []usecase.TeamMembers{
			{
				UserID:   userID,
				Username: "alice",
				IsActive: true,
			},
		},
	}

	tests := []struct {
		name        string
		query       string
		mock        func()
		wantCode    int
		wantError   string
		wantSuccess *handler.Team
	}{
		{
			name:  "success",
			query: "?team_name=" + teamName,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(&ucOut, nil)
			},
			wantCode: http.StatusOK,
			wantSuccess: &handler.Team{
				TeamName: teamName,
				Members: []handler.TeamMember{
					{
						UserId:   userID,
						Username: "alice",
						IsActive: true,
					},
				},
			},
		},
		{
			name:      "empty team_name",
			query:     "?team_name=",
			mock:      func() {},
			wantCode:  http.StatusBadRequest,
			wantError: "team_name cannot be empty",
		},
		{
			name:      "missing team_name parameter",
			query:     "",
			mock:      func() {},
			wantCode:  http.StatusBadRequest,
			wantError: "team_name cannot be empty",
		},
		{
			name:  "ErrGetTeam",
			query: "?team_name=" + teamName,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrGetTeam)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting team with such id",
		},
		{
			name:  "ErrGetUsers",
			query: "?team_name=" + teamName,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrGetUsers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while checking if users exist with such ids",
		},
		{
			name:  "ErrTeamNotFound",
			query: "?team_name=" + teamName,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrTeamNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "there is no team looking for",
		},
		{
			name:  "unknown error",
			query: "?team_name=" + teamName,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, fmt.Errorf("unknown error"))
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mock != nil {
				tt.mock()
			}

			req := httptest.NewRequest("GET", "/team/get"+tt.query, nil)
			w := httptest.NewRecorder()

			h.GetTeam(w, req)

			assert.Equal(t, tt.wantCode, w.Code, "Status code mismatch for test: %s", tt.name)

			if tt.wantSuccess != nil {
				var got handler.Team
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
				assert.Equal(t, *tt.wantSuccess, got, "Response body mismatch for test: %s", tt.name)
			}

			if tt.wantError != "" {
				var got struct {
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
				assert.Contains(t, got.Error.Message, tt.wantError, "Error message mismatch for test: %s", tt.name)
			}
		})
	}
}
