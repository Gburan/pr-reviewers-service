package stats_pr_assignments_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	handlerStats "pr-reviewers-service/internal/handler/stats_pr_assignments"
	mockStats "pr-reviewers-service/internal/handler/stats_pr_assignments/mocks"
	usecase2 "pr-reviewers-service/internal/usecase"
	usecaseStats "pr-reviewers-service/internal/usecase/stats_pr_assignments"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetReviewersStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mockStats.NewMockusecase(ctrl)
	h := handlerStats.New(mockUC)

	reviewerID := uuid.New()
	ucOut := usecaseStats.Out{
		Reviewers: []usecaseStats.ReviewerStats{
			{
				ReviewerID:      reviewerID,
				AssignmentCount: 5,
			},
		},
	}

	tests := []struct {
		name      string
		mock      func()
		wantCode  int
		wantError string
		wantBody  *handler2.ReviewersStatsResponse
	}{
		{
			name: "success",
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseStats.In{}).Return(&ucOut, nil)
			},
			wantCode: http.StatusOK,
			wantBody: &handler2.ReviewersStatsResponse{
				Reviewers: []handler2.ReviewerAssignmentCount{
					{
						ReviewerId:      reviewerID,
						AssignmentCount: 5,
					},
				},
			},
		},
		{
			name: "usecase returns ErrGetPRReviewers",
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseStats.In{}).Return(nil, usecase2.ErrGetPRReviewers)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "error occurred while getting PR reviewers",
		},
		{
			name: "usecase returns ErrPRsReviewersNotFound",
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseStats.In{}).Return(nil, usecase2.ErrPRsReviewersNotFound)
			},
			wantCode:  http.StatusNotFound,
			wantError: "PR reviewers not found",
		},
		{
			name: "usecase returns unknown error",
			mock: func() {
				mockUC.EXPECT().Run(gomock.Any(), usecaseStats.In{}).Return(nil, errors.New("unknown error"))
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

			req := httptest.NewRequest("GET", "/stats/reviewers", bytes.NewReader(nil))
			w := httptest.NewRecorder()

			h.GetReviewersStats(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantError != "" {
				var errResp struct {
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
				assert.Contains(t, errResp.Error.Message, tt.wantError)
			}

			if tt.wantBody != nil {
				var resp handler2.ReviewersStatsResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Equal(t, tt.wantBody, &resp)
			}
		})
	}
}
