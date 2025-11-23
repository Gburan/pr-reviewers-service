package pull_request_create_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	handlerPR "pr-reviewers-service/internal/handler/pull_request_create"
	mockPR "pr-reviewers-service/internal/handler/pull_request_create/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecase "pr-reviewers-service/internal/usecase/pull_request_create"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validate := validator.New()
	mockUC := mockPR.NewMockusecase(ctrl)
	h := handlerPR.New(mockUC, validate)

	prID := uuid.New()
	authorID := uuid.New()

	reqBody := handler.PostPullRequestCreateJSONRequestBody{
		PullRequestId:   prID,
		PullRequestName: "Add new feature",
		AuthorId:        authorID,
	}

	now := time.Now()
	assigned := []uuid.UUID{uuid.New(), uuid.New()}

	ucOut := usecase.Out{
		PullRequestID:     prID,
		PullRequestName:   "Add new feature",
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: assigned,
		CreatedAt:         now,
		MergedAt:          time.Time{},
	}

	tests := []struct {
		name        string
		body        interface{}
		mock        func()
		wantCode    int
		wantError   string
		wantSuccess *handler.CreatePullRequestResponse
	}{
		{
			name: "success",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(&ucOut, nil)
			},
			wantCode: http.StatusCreated,
			wantSuccess: &handler.CreatePullRequestResponse{
				Pr: handler.PullRequest{
					PullRequestId:     prID,
					PullRequestName:   "Add new feature",
					AuthorId:          authorID,
					Status:            handler.PullRequestStatus("OPEN"),
					AssignedReviewers: assigned,
					CreatedAt:         &now,
					MergedAt:          nil,
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
			name: "validation failed",
			body: map[string]interface{}{
				"pull_request_name": prID,
				"author_id":         authorID,
			},
			mock:      func() {},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "validation failed",
		},
		{
			name: "usecase returns ErrAuthorPrNotFound",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(nil, usecase2.ErrAuthorPrNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "author not found",
		},
		{
			name: "usecase returns ErrGetPullRequest",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(nil, usecase2.ErrGetPullRequest)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while checking pull request existence",
		},
		{
			name: "usecase returns ErrGetUser",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(nil, usecase2.ErrGetUser)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting author information",
		},
		{
			name: "usecase returns ErrGetUsers",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(nil, usecase2.ErrGetUsers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting team members",
		},
		{
			name: "usecase returns ErrSetPRStatus",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(nil, usecase2.ErrSetPRStatus)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while setting PR status",
		},
		{
			name: "usecase returns ErrSavePullRequest",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(nil, usecase2.ErrSavePullRequest)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while saving pull request in db",
		},
		{
			name: "usecase returns ErrAssignReviewer",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(nil, usecase2.ErrAssignReviewer)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while assigning reviewers",
		},
		{
			name: "usecase returns ErrPullRequestExists",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
				}).Return(nil, usecase2.ErrPullRequestExists)
			},
			wantCode:  http.StatusBadRequest,
			wantError: "pull request already exists",
		},
		{
			name: "usecase returns unknown error",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID:   prID,
					PullRequestName: "Add new feature",
					AuthorID:        authorID,
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

			req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.CreatePullRequest(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantSuccess != nil {
				var got handler.CreatePullRequestResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

				assert.Equal(t, tt.wantSuccess.Pr.PullRequestId, got.Pr.PullRequestId)
				assert.Equal(t, tt.wantSuccess.Pr.PullRequestName, got.Pr.PullRequestName)
				assert.Equal(t, tt.wantSuccess.Pr.AuthorId, got.Pr.AuthorId)
				assert.Equal(t, tt.wantSuccess.Pr.Status, got.Pr.Status)
				assert.Equal(t, tt.wantSuccess.Pr.AssignedReviewers, got.Pr.AssignedReviewers)
				assert.Nil(t, got.Pr.MergedAt)
				assert.NotNil(t, got.Pr.CreatedAt)
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
