package pull_request_reassign_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	handlerPR "pr-reviewers-service/internal/handler/pull_request_reassign"
	mockPR "pr-reviewers-service/internal/handler/pull_request_reassign/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecase "pr-reviewers-service/internal/usecase/pull_request_reassign"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReassignPullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validate := validator.New()
	mockUC := mockPR.NewMockusecase(ctrl)
	h := handlerPR.New(mockUC, validate)

	prID := uuid.New()
	oldReviewerID := uuid.New()
	newReviewerID := uuid.New()
	authorID := uuid.New()

	reqBody := handler.PostPullRequestReassignJSONRequestBody{
		PullRequestId: prID,
		OldReviewerId: oldReviewerID,
	}

	now := time.Now()
	assigned := []uuid.UUID{newReviewerID}

	ucOut := usecase.Out{
		PullRequestID:     prID,
		PullRequestName:   "Fix bug",
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: assigned,
		CreatedAt:         now,
		MergedAt:          time.Time{},
		ReplacedBy:        newReviewerID,
	}

	tests := []struct {
		name        string
		body        interface{}
		mock        func()
		wantCode    int
		wantError   string
		wantSuccess *handler.ReassignPullRequestResponse
	}{
		{
			name: "success",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(&ucOut, nil)
			},
			wantCode: http.StatusOK,
			wantSuccess: &handler.ReassignPullRequestResponse{
				Pr: handler.PullRequest{
					PullRequestId:     prID,
					PullRequestName:   "Fix bug",
					AuthorId:          authorID,
					Status:            handler.PullRequestStatus("OPEN"),
					AssignedReviewers: assigned,
					CreatedAt:         &now,
					MergedAt:          nil,
				},
				ReplacedBy: newReviewerID,
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
			name:      "validation failed",
			body:      map[string]interface{}{},
			mock:      func() {},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "validation failed",
		},
		{
			name: "usecase returns ErrPullRequestNotFound",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrPullRequestNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "pull request not found",
		},
		{
			name: "usecase returns ErrReviewerNotFound",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrReviewerNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "reviewer not found",
		},
		{
			name: "usecase returns ErrNoAvailableReviewers",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrNoAvailableReviewers)
			},
			wantCode:  http.StatusNotFound,
			wantError: "no available reviewers",
		},
		{
			name: "usecase returns ErrPullRequestAlreadyMerged",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrPullRequestAlreadyMerged)
			},
			wantCode:  http.StatusConflict,
			wantError: "pull request already merged",
		},
		{
			name: "usecase returns ErrAuthorPrNotFound",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
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
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrGetPullRequest)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pull request",
		},
		{
			name: "usecase returns ErrGetPRStatus",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrGetPRStatus)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pr status",
		},
		{
			name: "usecase returns ErrGetPRReviewers",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrGetPRReviewers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pr reviewers",
		},
		{
			name: "usecase returns ErrGetUser",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrGetUser)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting user",
		},
		{
			name: "usecase returns ErrGetUsers",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrGetUsers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting users",
		},
		{
			name: "usecase returns ErrRemoveReviewer",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrRemoveReviewer)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while removing reviewer",
		},
		{
			name: "usecase returns ErrAssignReviewer",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
				}).Return(nil, usecase2.ErrAssignReviewer)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while assigning reviewer",
		},
		{
			name: "usecase returns unknown error",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
					OldUserId:     oldReviewerID,
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

			req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.ReassignPullRequest(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantSuccess != nil {
				var got handler.ReassignPullRequestResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

				assert.Equal(t, tt.wantSuccess.Pr.PullRequestId, got.Pr.PullRequestId)
				assert.Equal(t, tt.wantSuccess.Pr.PullRequestName, got.Pr.PullRequestName)
				assert.Equal(t, tt.wantSuccess.Pr.AuthorId, got.Pr.AuthorId)
				assert.Equal(t, tt.wantSuccess.Pr.Status, got.Pr.Status)
				assert.Equal(t, tt.wantSuccess.Pr.AssignedReviewers, got.Pr.AssignedReviewers)
				assert.Equal(t, tt.wantSuccess.ReplacedBy, got.ReplacedBy)
				assert.NotNil(t, got.Pr.CreatedAt)
				assert.Nil(t, got.Pr.MergedAt)
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
