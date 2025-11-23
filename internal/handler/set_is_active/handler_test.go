package set_is_active_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	handlerSet "pr-reviewers-service/internal/handler/set_is_active"
	mockSet "pr-reviewers-service/internal/handler/set_is_active/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecase "pr-reviewers-service/internal/usecase/set_is_active"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetIsActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validate := validator.New()
	mockUC := mockSet.NewMockusecase(ctrl)
	h := handlerSet.New(mockUC, validate)

	userID := uuid.New()
	reqBody := handler.PostUsersSetIsActiveJSONRequestBody{
		UserId:   userID,
		IsActive: true,
	}

	ucOut := usecase.Out{
		UserId:   userID,
		Username: "user1",
		TeamName: "teamA",
		IsActive: true,
	}

	tests := []struct {
		name        string
		body        interface{}
		mock        func()
		wantCode    int
		wantError   string
		wantSuccess *handler.SetUserActiveStatusResponse
	}{
		{
			name: "success",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					UserID:   userID,
					IsActive: true,
				}).Return(&ucOut, nil)
			},
			wantCode: http.StatusOK,
			wantSuccess: &handler.SetUserActiveStatusResponse{
				User: handler.User{
					UserId:   userID,
					Username: "user1",
					TeamName: "teamA",
					IsActive: true,
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
			name: "usecase returns ErrUserNotFound",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					UserID:   userID,
					IsActive: true,
				}).Return(nil, usecase2.ErrUserNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "user not found",
		},
		{
			name: "usecase returns ErrUpdateUser",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					UserID:   userID,
					IsActive: true,
				}).Return(nil, usecase2.ErrUpdateUser)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while updating user in db",
		},
		{
			name: "usecase returns ErrGetUser",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					UserID:   userID,
					IsActive: true,
				}).Return(nil, usecase2.ErrGetUser)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting user from db",
		},
		{
			name: "usecase returns ErrGetTeam",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					UserID:   userID,
					IsActive: true,
				}).Return(nil, usecase2.ErrGetTeam)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting users team",
		},
		{
			name: "usecase returns ErrUserDontNeedChange",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					UserID:   userID,
					IsActive: true,
				}).Return(nil, usecase2.ErrUserDontNeedChange)
			},
			wantCode:  http.StatusNotModified,
			wantError: "user already have same status as you trying to assign",
		},
		{
			name: "usecase returns unknown error",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					UserID:   userID,
					IsActive: true,
				}).Return(nil, errors.New("unknown error"))
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
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.SetIsActive(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantSuccess != nil {
				var got handler.SetUserActiveStatusResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
				assert.Equal(t, *tt.wantSuccess, got)
			}

			if tt.wantError != "" {
				var errResp struct {
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
				assert.Contains(t, errResp.Error.Message, tt.wantError)
			}
		})
	}
}
