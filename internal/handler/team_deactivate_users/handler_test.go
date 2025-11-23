package team_deactivate_users_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	handlerTeam "pr-reviewers-service/internal/handler/team_deactivate_users"
	mockTeam "pr-reviewers-service/internal/handler/team_deactivate_users/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecaseTeam "pr-reviewers-service/internal/usecase/team_deactivate_users"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeactivateTeamUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validate := validator.New()
	mockUC := mockTeam.NewMockusecase(ctrl)
	h := handlerTeam.New(mockUC, validate)

	u1 := uuid.New()
	u2 := uuid.New()
	u3 := uuid.New()
	pr1 := uuid.New()

	reqBody := handler2.PatchTeamDeactivateUsersJSONRequestBody{
		TeamName: "teamA",
		UserIds:  []uuid.UUID{u1, u2},
	}

	ucOut := usecaseTeam.Out{
		Team: usecaseTeam.Team{
			TeamName: "teamA",
			Members: []usecaseTeam.TeamMember{
				{UserID: u1, Username: "user1", IsActive: false},
				{UserID: u2, Username: "user2", IsActive: false},
			},
		},
		AffectedPullRequests: []usecaseTeam.PullRequestShort{
			{PullRequestID: pr1, PullRequestName: "PR1", AuthorID: u3, Status: "OPEN"},
		},
	}

	tests := []struct {
		name      string
		body      interface{}
		mock      func()
		wantCode  int
		wantError string
		wantBody  interface{}
	}{
		{
			name: "success",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(&ucOut, nil)
			},
			wantCode: http.StatusOK,
			wantBody: struct {
				Team                 handler2.Team               `json:"team"`
				AffectedPullRequests []handler2.PullRequestShort `json:"affected_pull_requests"`
			}{
				Team: handler2.Team{
					TeamName: "teamA",
					Members: []handler2.TeamMember{
						{UserId: u1, Username: "user1", IsActive: false},
						{UserId: u2, Username: "user2", IsActive: false},
					},
				},
				AffectedPullRequests: []handler2.PullRequestShort{
					{PullRequestId: pr1, PullRequestName: "PR1", AuthorId: u3, Status: handler2.PullRequestShortStatus("OPEN")},
				},
			},
		},
		{
			name:      "invalid JSON",
			body:      "invalid-json",
			mock:      func() {},
			wantCode:  http.StatusBadRequest,
			wantError: "failed to decode request",
		},
		{
			name:      "empty body",
			body:      nil,
			mock:      func() {},
			wantCode:  http.StatusBadRequest,
			wantError: "failed to decode request",
		},
		{
			name: "validation failed - missing team name",
			body: struct {
				UserIds []uuid.UUID `json:"user_ids"`
			}{
				UserIds: []uuid.UUID{u1, u2},
			},
			mock:      func() {},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "validation failed",
		},
		{
			name: "validation failed - missing user IDs",
			body: struct {
				TeamName string `json:"team_name"`
			}{
				TeamName: "teamA",
			},
			mock:      func() {},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "validation failed",
		},
		{
			name: "validation failed - empty user IDs",
			body: struct {
				TeamName string      `json:"team_name"`
				UserIds  []uuid.UUID `json:"user_ids"`
			}{
				TeamName: "teamA",
				UserIds:  []uuid.UUID{},
			},
			mock:      func() {},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "validation failed",
		},
		{
			name: "usecase returns ErrTeamNotFound",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrTeamNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "team not found",
		},
		{
			name: "usecase returns ErrUserNotBelongsToTeam",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrUserNotBelongsToTeam)
			},
			wantCode:  http.StatusConflict,
			wantError: "user does not belong to the specified team",
		},
		{
			name: "usecase returns ErrUserNotFound",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrUserNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "user not found",
		},
		{
			name: "usecase returns ErrUsersByIDsNotFound",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrUsersByIDsNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "users not found by provided IDs",
		},
		{
			name: "usecase returns ErrNoPRsToAffect",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrNoPRsToAffect)
			},
			wantCode:  http.StatusNotFound,
			wantError: "no pull requests to affect",
		},
		{
			name: "usecase returns ErrNoUsersAssignedToPRs",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrNoUsersAssignedToPRs)
			},
			wantCode:  http.StatusNotFound,
			wantError: "no users assigned to pull requests",
		},
		{
			name: "usecase returns ErrGetTeam",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrGetTeam)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting team",
		},
		{
			name: "usecase returns ErrGetUsers",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrGetUsers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting users",
		},
		{
			name: "usecase returns ErrUpdateUser",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrUpdateUser)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while updating user",
		},
		{
			name: "usecase returns ErrGetPRReviewers",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrGetPRReviewers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pr reviewers",
		},
		{
			name: "usecase returns ErrGetPullRequest",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrGetPullRequest)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pull request",
		},
		{
			name: "usecase returns ErrGetPRStatus",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrGetPRStatus)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pr status",
		},
		{
			name: "usecase returns ErrRemoveReviewer",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrRemoveReviewer)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while removing reviewer",
		},
		{
			name: "usecase returns ErrAssignReviewer",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, usecase2.ErrAssignReviewer)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while assigning reviewer",
		},
		{
			name: "usecase returns unknown error",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(nil, errors.New("unknown error"))
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "internal server error",
		},
		{
			name: "json encode error",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseTeam.In{
					TeamName: "teamA",
					UserIDs:  []uuid.UUID{u1, u2},
				}).Return(&ucOut, nil)
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mock != nil {
				tt.mock()
			}

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			case nil:
				bodyBytes = nil
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("PATCH", "/team/deactivateUsers", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.DeactivateTeamUsers(w, req)

			assert.Equal(t, tt.wantCode, w.Code, "Status code mismatch for test: %s", tt.name)

			if tt.wantError != "" {
				var errResp struct {
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
				assert.Contains(t, errResp.Error.Message, tt.wantError, "Error message mismatch for test: %s", tt.name)
			}

			if tt.wantBody != nil {
				var resp struct {
					Team                 handler2.Team               `json:"team"`
					AffectedPullRequests []handler2.PullRequestShort `json:"affected_pull_requests"`
				}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Equal(t, tt.wantBody, resp, "Response body mismatch for test: %s", tt.name)
			}
		})
	}
}
