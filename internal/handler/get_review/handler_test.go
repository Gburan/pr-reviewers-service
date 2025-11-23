package get_review_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "pr-reviewers-service/internal/generated/api/v1/handler"
	get_review_handler "pr-reviewers-service/internal/handler/get_review"
	mock_review "pr-reviewers-service/internal/handler/get_review/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecase "pr-reviewers-service/internal/usecase/get_review"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserReviewPRs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validate := validator.New()

	mockUC := mock_review.NewMockusecase(ctrl)
	h := get_review_handler.New(mockUC, validate)

	u1 := uuid.New()
	uAuthor := uuid.New()
	prID := uuid.New()

	ucIn := usecase.In{UserID: u1}

	ucOut := usecase.Out{
		UserID: u1,
		PullRequests: []usecase.PullRequestShort{
			{
				PullRequestID:   prID,
				PullRequestName: "Improve API",
				AuthorID:        uAuthor,
				Status:          "open",
			},
		},
	}

	tests := []struct {
		name        string
		query       string
		mock        func()
		wantCode    int
		wantError   string
		wantSuccess *handler.GetUserReviewPRsResponse
	}{
		{
			name:  "success",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(&ucOut, nil)
			},
			wantCode: http.StatusOK,
			wantSuccess: &handler.GetUserReviewPRsResponse{
				UserId: u1,
				PullRequests: []handler.PullRequestShort{
					{
						PullRequestId:   prID,
						PullRequestName: "Improve API",
						AuthorId:        uAuthor,
						Status:          handler.PullRequestShortStatus("open"),
					},
				},
			},
		},
		{
			name:      "missing user_id",
			query:     "",
			mock:      func() {},
			wantCode:  http.StatusBadRequest,
			wantError: "user_id is required",
		},
		{
			name:      "invalid uuid",
			query:     "?user_id=abc",
			mock:      func() {},
			wantCode:  http.StatusBadRequest,
			wantError: "invalid user_id format",
		},
		{
			name:  "usecase returns ErrGetPRReviewers",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrGetPRReviewers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pr ids where user reviewer",
		},
		{
			name:  "usecase returns ErrGetPullRequest",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrGetPullRequest)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pull requests",
		},
		{
			name:  "usecase returns ErrGetUser",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrGetUser)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting user from db",
		},
		{
			name:  "usecase returns ErrGetPRStatus",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrGetPRStatus)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pr statuses",
		},
		{
			name:  "usecase returns ErrUserNotFound",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrUserNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "user not found",
		},
		{
			name:  "usecase returns ErrNoActiveReviewers",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, usecase2.ErrNoActiveReviewers)
			},
			wantCode:  http.StatusNotFound,
			wantError: "user is not assigned as reviewer to any pull requests",
		},
		{
			name:  "usecase returns unknown error",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, fmt.Errorf("unknown error"))
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "internal server error",
		},
		{
			name:  "usecase returns assert error",
			query: "?user_id=" + u1.String(),
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), ucIn).Return(nil, assert.AnError)
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

			req := httptest.NewRequest("GET", "/review"+tt.query, nil)
			w := httptest.NewRecorder()

			h.GetUserReviewPRs(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantSuccess != nil {
				var got handler.GetUserReviewPRsResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
				assert.Equal(t, *tt.wantSuccess, got)
				return
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
