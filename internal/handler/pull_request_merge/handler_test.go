package pull_request_merge_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	handlerPR "pr-reviewers-service/internal/handler/pull_request_merge"
	mockPR "pr-reviewers-service/internal/handler/pull_request_merge/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecase "pr-reviewers-service/internal/usecase/pull_request_merge"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergePullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validate := validator.New()
	mockUC := mockPR.NewMockusecase(ctrl)
	h := handlerPR.New(mockUC, validate)

	prID := uuid.New()
	authorID := uuid.New()

	reqBody := handler.PostPullRequestMergeJSONRequestBody{
		PullRequestId: prID,
	}

	now := time.Now()
	assigned := []uuid.UUID{uuid.New(), uuid.New()}

	ucOut := usecase.Out{
		PullRequestID:     prID,
		PullRequestName:   "Fix bug",
		AuthorID:          authorID,
		Status:            "MERGED",
		AssignedReviewers: assigned,
		CreatedAt:         now,
		MergedAt:          now,
	}

	tests := []struct {
		name        string
		body        interface{}
		mock        func()
		wantCode    int
		wantError   string
		wantSuccess *handler.MergePullRequestResponse
	}{
		{
			name: "success",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
				}).Return(&ucOut, nil)
			},
			wantCode: http.StatusOK,
			wantSuccess: &handler.MergePullRequestResponse{
				Pr: handler.PullRequest{
					PullRequestId:     prID,
					PullRequestName:   "Fix bug",
					AuthorId:          authorID,
					Status:            handler.PullRequestStatus("MERGED"),
					AssignedReviewers: assigned,
					CreatedAt:         &now,
					MergedAt:          &now,
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
				}).Return(nil, usecase2.ErrPullRequestNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "pull request not found",
		},
		{
			name: "usecase returns ErrGetPullRequest",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
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
				}).Return(nil, usecase2.ErrGetPRReviewers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting pr reviewers",
		},
		{
			name: "usecase returns ErrSetPRStatus",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
				}).Return(nil, usecase2.ErrSetPRStatus)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while setting pr status",
		},
		{
			name: "usecase returns ErrUpdatePrMergeTime",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
				}).Return(nil, usecase2.ErrUpdatePrMergeTime)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while updating pr merge time",
		},
		{
			name: "usecase returns unknown error",
			body: reqBody,
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecase.In{
					PullRequestID: prID,
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

			req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.MergePullRequest(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantSuccess != nil {
				var got handler.MergePullRequestResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&got))

				assert.Equal(t, tt.wantSuccess.Pr.PullRequestId, got.Pr.PullRequestId)
				assert.Equal(t, tt.wantSuccess.Pr.PullRequestName, got.Pr.PullRequestName)
				assert.Equal(t, tt.wantSuccess.Pr.AuthorId, got.Pr.AuthorId)
				assert.Equal(t, tt.wantSuccess.Pr.Status, got.Pr.Status)
				assert.Equal(t, tt.wantSuccess.Pr.AssignedReviewers, got.Pr.AssignedReviewers)
				assert.NotNil(t, got.Pr.CreatedAt)
				assert.NotNil(t, got.Pr.MergedAt)
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
