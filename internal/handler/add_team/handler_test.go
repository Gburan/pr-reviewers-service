package add_team_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	add_team_handler "pr-reviewers-service/internal/handler/add_team"
	mock_add_team "pr-reviewers-service/internal/handler/add_team/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecase "pr-reviewers-service/internal/usecase/add_team"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validate := validator.New()
	mockUC := mock_add_team.NewMockusecase(ctrl)
	h := add_team_handler.New(mockUC, validate)

	userID := uuid.New()

	reqBody := handler.PostTeamAddJSONRequestBody{
		TeamName: "backend",
		Members: []handler.TeamMember{
			{
				UserId:   userID,
				Username: "alice",
				IsActive: true,
			},
		},
	}

	ucIn := usecase.In{
		TeamName: reqBody.TeamName,
		Members: []usecase.TeamMembers{
			{
				UserID:   userID,
				Username: "alice",
				IsActive: true,
			},
		},
	}

	ucOut := usecase.Out{
		TeamName: "backend",
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
		reqBody     interface{}
		mock        func()
		wantCode    int
		wantError   string
		wantSuccess *handler.Team // Теперь ожидаем напрямую Team, а не обертку
	}{
		{
			name:    "success",
			reqBody: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(&ucOut, nil)
			},
			wantCode: http.StatusCreated,
			wantSuccess: &handler.Team{
				TeamName: "backend",
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
			name:      "decode error",
			reqBody:   "not json",
			mock:      func() {},
			wantCode:  http.StatusBadRequest,
			wantError: "failed to decode request",
		},
		{
			name: "validation error",
			reqBody: handler.PostTeamAddJSONRequestBody{
				TeamName: "",
				Members:  []handler.TeamMember{},
			},
			mock:      func() {},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "validation failed",
		},
		{
			name:    "ErrNoUsersWereUpdatedAddedTeam",
			reqBody: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrNoUsersWereUpdatedAddedTeam)
			},
			wantCode:  http.StatusNotModified,
			wantError: "team exists and no users were changed or added",
		},
		{
			name:    "ErrDuplicateUsers",
			reqBody: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrDuplicateUsers)
			},
			wantCode:  http.StatusBadRequest,
			wantError: "dont use same user ids",
		},
		{
			name:    "ErrGetTeam",
			reqBody: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrGetTeam)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting team with such id",
		},
		{
			name:    "ErrSaveTeam",
			reqBody: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrSaveTeam)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while saving team in db",
		},
		{
			name:    "ErrGetUsers",
			reqBody: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrGetUsers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while checking if users exist with such ids",
		},
		{
			name:    "ErrSaveUsersBatch",
			reqBody: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrSaveUsersBatch)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while saving new users in db",
		},
		{
			name:    "ErrUpdateUsersBatch",
			reqBody: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrUpdateUsersBatch)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while updating existing users in db",
		},
		{
			name:    "unknown error",
			reqBody: reqBody,
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

			var bodyBytes []byte
			switch v := tt.reqBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.AddTeam(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantSuccess != nil {
				// Теперь декодируем напрямую в handler.Team, а не в обертку
				var got handler.Team
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
				assert.Equal(t, *tt.wantSuccess, got)
			}

			if tt.wantError != "" {
				var got struct {
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
				assert.Contains(t, got.Error.Message, tt.wantError)
			}
		})
	}
}
