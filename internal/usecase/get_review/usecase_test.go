package get_review

import (
	"context"
	"errors"
	"testing"

	"pr-reviewers-service/internal/infrastructure/repository"
	pr_reviewers2 "pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	pr_statuses2 "pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	pull_requests2 "pr-reviewers-service/internal/infrastructure/repository/pull_requests"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	usecase2 "pr-reviewers-service/internal/usecase"
	pr_reviewers "pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers/mocks"
	pr_statuses "pr-reviewers-service/internal/usecase/contract/repository/pr_statuses/mocks"
	pull_requests "pr-reviewers-service/internal/usecase/contract/repository/pull_requests/mocks"
	users "pr-reviewers-service/internal/usecase/contract/repository/users/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetReview(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	statusID1 := uuid.New()
	statusID2 := uuid.New()

	req := In{
		UserID: userID,
	}
	userOut := &users2.UserOut{
		ID:       userID,
		Name:     "test-user",
		IsActive: true,
		TeamID:   uuid.New(),
	}
	prReviewers := []pr_reviewers2.PrReviewerOut{
		{
			ID:         uuid.New(),
			PRID:       prID1,
			ReviewerID: userID,
		},
		{
			ID:         uuid.New(),
			PRID:       prID2,
			ReviewerID: userID,
		},
	}
	pullRequests := []pull_requests2.PullRequestOut{
		{
			ID:       prID1,
			Name:     "PR-1",
			AuthorID: uuid.New(),
			StatusID: statusID1,
		},
		{
			ID:       prID2,
			Name:     "PR-2",
			AuthorID: uuid.New(),
			StatusID: statusID2,
		},
	}
	prStatuses := []pr_statuses2.PRStatusOut{
		{
			ID:     statusID1,
			Status: "open",
		},
		{
			ID:     statusID2,
			Status: "closed",
		},
	}

	tests := []struct {
		name      string
		req       In
		setupMock func(
			mockUsers *users.MockRepositoryUsers,
			mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
			mockPullRequests *pull_requests.MockRepositoryPullRequests,
			mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
		)
		expected      *Out
		expectedError error
	}{
		{
			name: "successful get reviews",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(&prReviewers, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{prID1, prID2}).
					Return(&pullRequests, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID1, statusID2}).
					Return(&prStatuses, nil)
			},
			expected: &Out{
				UserID: userID,
				PullRequests: []PullRequestShort{
					{
						PullRequestID:   prID1,
						PullRequestName: "PR-1",
						AuthorID:        pullRequests[0].AuthorID,
						Status:          "open",
					},
					{
						PullRequestID:   prID2,
						PullRequestName: "PR-2",
						AuthorID:        pullRequests[1].AuthorID,
						Status:          "closed",
					},
				},
			},
		},
		{
			name: "user not found",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(nil, repository.ErrUserNotFound)
			},
			expectedError: usecase2.ErrUserNotFound,
		},
		{
			name: "error getting user",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(nil, errors.New("database error"))
			},
			expectedError: usecase2.ErrGetUser,
		},
		{
			name: "no PR reviewers found",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(nil, repository.ErrPRReviewerNotFound)
			},
			expectedError: usecase2.ErrNoActiveReviewers,
		},
		{
			name: "error getting PR reviewers",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(nil, errors.New("database error"))
			},
			expectedError: usecase2.ErrGetPRReviewers,
		},
		{
			name: "no pull requests found",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(&prReviewers, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{prID1, prID2}).
					Return(nil, repository.ErrPullRequestNotFound)
			},
			expected: &Out{
				UserID:       userID,
				PullRequests: []PullRequestShort{},
			},
		},
		{
			name: "error getting pull requests",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(&prReviewers, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{prID1, prID2}).
					Return(nil, errors.New("database error"))
			},
			expectedError: usecase2.ErrGetPullRequest,
		},
		{
			name: "error getting PR statuses",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(&prReviewers, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{prID1, prID2}).
					Return(&pullRequests, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID1, statusID2}).
					Return(nil, errors.New("database error"))
			},
			expectedError: usecase2.ErrGetPRStatus,
		},
		{
			name: "empty PR statuses returned",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(&prReviewers, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{prID1, prID2}).
					Return(&pullRequests, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID1, statusID2}).
					Return(&[]pr_statuses2.PRStatusOut{}, nil)
			},
			expected: &Out{
				UserID:       userID,
				PullRequests: []PullRequestShort{},
			},
		},
		{
			name: "partial PR statuses found - skip PRs without status",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(&prReviewers, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{prID1, prID2}).
					Return(&pullRequests, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID1, statusID2}).
					Return(&[]pr_statuses2.PRStatusOut{
						{
							ID:     statusID1,
							Status: "open",
						},
					}, nil)
			},
			expected: &Out{
				UserID: userID,
				PullRequests: []PullRequestShort{
					{
						PullRequestID:   prID1,
						PullRequestName: "PR-1",
						AuthorID:        pullRequests[0].AuthorID,
						Status:          "open",
					},
				},
			},
		},
		{
			name: "empty PR reviewers list",
			req:  req,
			setupMock: func(
				mockUsers *users.MockRepositoryUsers,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(userOut, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerID(gomock.Any(), userID).
					Return(&[]pr_reviewers2.PrReviewerOut{}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{}).
					Return(nil, repository.ErrPullRequestNotFound)
			},
			expected: &Out{
				UserID:       userID,
				PullRequests: []PullRequestShort{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoUsers := users.NewMockRepositoryUsers(ctrl)
			mockRepoPRReviewers := pr_reviewers.NewMockRepositoryPrReviewers(ctrl)
			mockRepoPullRequests := pull_requests.NewMockRepositoryPullRequests(ctrl)
			mockRepoPRStatuses := pr_statuses.NewMockRepositoryPrStatuses(ctrl)

			tt.setupMock(mockRepoUsers, mockRepoPRReviewers, mockRepoPullRequests, mockRepoPRStatuses)

			u := NewUsecase(
				mockRepoUsers,
				mockRepoPullRequests,
				mockRepoPRReviewers,
				mockRepoPRStatuses,
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
				assert.Equal(t, tt.expected.UserID, result.UserID)
				assert.Equal(t, len(tt.expected.PullRequests), len(result.PullRequests))

				for i, expectedPR := range tt.expected.PullRequests {
					assert.Equal(t, expectedPR.PullRequestID, result.PullRequests[i].PullRequestID)
					assert.Equal(t, expectedPR.PullRequestName, result.PullRequests[i].PullRequestName)
					assert.Equal(t, expectedPR.AuthorID, result.PullRequests[i].AuthorID)
					assert.Equal(t, expectedPR.Status, result.PullRequests[i].Status)
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
