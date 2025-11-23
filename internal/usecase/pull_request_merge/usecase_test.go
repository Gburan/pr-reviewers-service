package pull_request_merge

import (
	"context"
	"errors"
	"testing"
	"time"

	"pr-reviewers-service/internal/infrastructure/repository"
	pr_reviewers2 "pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	pr_statuses2 "pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	pull_requests2 "pr-reviewers-service/internal/infrastructure/repository/pull_requests"
	usecase2 "pr-reviewers-service/internal/usecase"
	pr_reviewers "pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers/mocks"
	pr_statuses "pr-reviewers-service/internal/usecase/contract/repository/pr_statuses/mocks"
	pull_requests "pr-reviewers-service/internal/usecase/contract/repository/pull_requests/mocks"

	"github.com/avito-tech/go-transaction-manager/trm/v2/drivers/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullRequestMerge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prID := uuid.New()
	authorID := uuid.New()
	statusID := uuid.New()
	reviewerID1 := uuid.New()
	reviewerID2 := uuid.New()

	req := In{
		PullRequestID: prID,
	}
	existingPR := &pull_requests2.PullRequestOut{
		ID:        prID,
		Name:      "Test PR",
		AuthorID:  authorID,
		StatusID:  statusID,
		CreatedAt: time.Now(),
	}
	currentStatus := &pr_statuses2.PRStatusOut{
		ID:     statusID,
		Status: "OPEN",
	}
	mergedStatus := &pr_statuses2.PRStatusOut{
		ID:     statusID,
		Status: usecase2.MergedStatusValue,
	}

	reviewers := []pr_reviewers2.PrReviewerOut{
		{
			ID:         uuid.New(),
			PRID:       prID,
			ReviewerID: reviewerID1,
		},
		{
			ID:         uuid.New(),
			PRID:       prID,
			ReviewerID: reviewerID2,
		},
	}

	updatedPR := &pull_requests2.PullRequestOut{
		ID:        prID,
		Name:      "Test PR",
		AuthorID:  authorID,
		StatusID:  statusID,
		CreatedAt: existingPR.CreatedAt,
		MergedAt:  time.Now().Add(time.Second),
	}

	tests := []struct {
		name      string
		req       In
		setupMock func(
			mockPullRequests *pull_requests.MockRepositoryPullRequests,
			mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
			mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			mockTrm *mock.MockManager,
		)
		expected      *Out
		expectedError error
	}{
		{
			name: "successful merge pull request",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(currentStatus, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&reviewers, nil)

				mockPRStatuses.EXPECT().
					UpdatePRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{
						ID:     statusID,
						Status: usecase2.MergedStatusValue,
					}).
					Return(mergedStatus, nil)

				mockPullRequests.EXPECT().
					MarkPullRequestMergedByID(gomock.Any(), prID).
					Return(updatedPR, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				PullRequestID:   prID,
				PullRequestName: "Test PR",
				AuthorID:        authorID,
				Status:          usecase2.MergedStatusValue,
				AssignedReviewers: []uuid.UUID{
					reviewerID1,
					reviewerID2,
				},
				CreatedAt: existingPR.CreatedAt,
				MergedAt:  updatedPR.MergedAt,
			},
		},
		{
			name: "pull request not found",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrPullRequestNotFound,
		},
		{
			name: "error getting pull request",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrGetPullRequest,
		},
		{
			name: "error getting PR status",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrGetPRStatus,
		},
		{
			name: "error getting PR reviewers",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(currentStatus, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrGetPRReviewers,
		},
		{
			name: "pull request already merged",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mergedStatus := &pr_statuses2.PRStatusOut{
					ID:     statusID,
					Status: usecase2.MergedStatusValue,
				}
				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(mergedStatus, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&reviewers, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				PullRequestID:   prID,
				PullRequestName: "Test PR",
				AuthorID:        authorID,
				Status:          usecase2.MergedStatusValue,
				AssignedReviewers: []uuid.UUID{
					reviewerID1,
					reviewerID2,
				},
				CreatedAt: existingPR.CreatedAt,
				MergedAt:  existingPR.MergedAt,
			},
			expectedError: usecase2.ErrPullRequestAlreadyMerged,
		},
		{
			name: "no reviewers found - empty list",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(currentStatus, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(nil, repository.ErrPRReviewerNotFound)

				mockPRStatuses.EXPECT().
					UpdatePRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{
						ID:     statusID,
						Status: usecase2.MergedStatusValue,
					}).
					Return(mergedStatus, nil)

				mockPullRequests.EXPECT().
					MarkPullRequestMergedByID(gomock.Any(), prID).
					Return(updatedPR, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				PullRequestID:     prID,
				PullRequestName:   "Test PR",
				AuthorID:          authorID,
				Status:            usecase2.MergedStatusValue,
				AssignedReviewers: []uuid.UUID{},
				CreatedAt:         existingPR.CreatedAt,
				MergedAt:          updatedPR.MergedAt,
			},
		},
		{
			name: "error updating PR status",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(currentStatus, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&reviewers, nil)

				mockPRStatuses.EXPECT().
					UpdatePRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{
						ID:     statusID,
						Status: usecase2.MergedStatusValue,
					}).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrUpdatePrStatus,
		},
		{
			name: "error updating pull request merge time",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(currentStatus, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&reviewers, nil)

				mockPRStatuses.EXPECT().
					UpdatePRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{
						ID:     statusID,
						Status: usecase2.MergedStatusValue,
					}).
					Return(mergedStatus, nil)

				mockPullRequests.EXPECT().
					MarkPullRequestMergedByID(gomock.Any(), prID).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrUpdatePrMergeTime,
		},
		{
			name: "successful merge with empty reviewers list",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(currentStatus, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&[]pr_reviewers2.PrReviewerOut{}, nil)

				mockPRStatuses.EXPECT().
					UpdatePRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{
						ID:     statusID,
						Status: usecase2.MergedStatusValue,
					}).
					Return(mergedStatus, nil)

				mockPullRequests.EXPECT().
					MarkPullRequestMergedByID(gomock.Any(), prID).
					Return(updatedPR, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				PullRequestID:     prID,
				PullRequestName:   "Test PR",
				AuthorID:          authorID,
				Status:            usecase2.MergedStatusValue,
				AssignedReviewers: []uuid.UUID{},
				CreatedAt:         existingPR.CreatedAt,
				MergedAt:          updatedPR.MergedAt,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoPullRequests := pull_requests.NewMockRepositoryPullRequests(ctrl)
			mockRepoPRReviewers := pr_reviewers.NewMockRepositoryPrReviewers(ctrl)
			mockRepoPRStatuses := pr_statuses.NewMockRepositoryPrStatuses(ctrl)
			mockTrm := mock.NewMockManager(ctrl)

			tt.setupMock(
				mockRepoPullRequests,
				mockRepoPRReviewers,
				mockRepoPRStatuses,
				mockTrm,
			)

			u := NewUsecase(
				mockRepoPullRequests,
				mockRepoPRReviewers,
				mockRepoPRStatuses,
				mockTrm,
			)

			result, err := u.Run(context.Background(), tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.expected != nil {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.PullRequestID, result.PullRequestID)
				assert.Equal(t, tt.expected.PullRequestName, result.PullRequestName)
				assert.Equal(t, tt.expected.AuthorID, result.AuthorID)
				assert.Equal(t, tt.expected.Status, result.Status)
				assert.Equal(t, len(tt.expected.AssignedReviewers), len(result.AssignedReviewers))
				assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
				assert.Equal(t, tt.expected.MergedAt, result.MergedAt)

				for i, expectedReviewer := range tt.expected.AssignedReviewers {
					assert.Equal(t, expectedReviewer, result.AssignedReviewers[i])
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
