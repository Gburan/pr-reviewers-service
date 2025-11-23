package pull_request_create

import (
	"context"
	"errors"
	"testing"
	"time"

	"pr-reviewers-service/internal/infrastructure/repository"
	pr_reviewers2 "pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	pr_statuses2 "pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	pull_requests2 "pr-reviewers-service/internal/infrastructure/repository/pull_requests"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	usecase2 "pr-reviewers-service/internal/usecase"
	randomizer "pr-reviewers-service/internal/usecase/contract/randomizer/mocks"
	pr_reviewers "pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers/mocks"
	pr_statuses "pr-reviewers-service/internal/usecase/contract/repository/pr_statuses/mocks"
	pull_requests "pr-reviewers-service/internal/usecase/contract/repository/pull_requests/mocks"
	users "pr-reviewers-service/internal/usecase/contract/repository/users/mocks"

	"github.com/avito-tech/go-transaction-manager/trm/v2/drivers/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	cntReviewers = 2
)

func TestPullRequestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prID := uuid.New()
	authorID := uuid.New()
	teamID := uuid.New()
	statusID := uuid.New()
	reviewerID1 := uuid.New()
	reviewerID2 := uuid.New()
	reviewerID3 := uuid.New()

	req := In{
		PullRequestID:   prID,
		PullRequestName: "Test PR",
		AuthorID:        authorID,
	}
	author := &users2.UserOut{
		ID:       authorID,
		Name:     "author",
		IsActive: true,
		TeamID:   teamID,
	}
	teamMembers := []users2.UserOut{
		{
			ID:       reviewerID1,
			Name:     "reviewer1",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       reviewerID2,
			Name:     "reviewer2",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       reviewerID3,
			Name:     "reviewer3",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       authorID,
			Name:     "author",
			IsActive: true,
			TeamID:   teamID,
		},
	}
	prStatusOut := &pr_statuses2.PRStatusOut{
		ID:     statusID,
		Status: "OPEN",
	}
	createdPR := &pull_requests2.PullRequestOut{
		ID:        prID,
		Name:      req.PullRequestName,
		AuthorID:  authorID,
		StatusID:  statusID,
		CreatedAt: time.Now(),
	}
	tests := []struct {
		name      string
		req       In
		setupMock func(
			mockPullRequests *pull_requests.MockRepositoryPullRequests,
			mockUsers *users.MockRepositoryUsers,
			mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
			mockRandomizer *randomizer.MockRandomizer,
			mockTrm *mock.MockManager,
		)
		expected      *Out
		expectedError error
	}{
		{
			name: "successful create pull request with reviewers",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockPRStatuses.EXPECT().
					SavePRStatus(gomock.Any(), gomock.Any()).
					Return(prStatusOut, nil)

				mockPullRequests.EXPECT().
					SavePullRequest(gomock.Any(), gomock.Any()).
					Return(createdPR, nil)

				mockRandomizer.EXPECT().
					Shuffle(3, gomock.Any()).
					DoAndReturn(func(n int, swap func(i, j int)) {
						swap(0, 1)
					})

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.Any()).
					Times(cntReviewers).Return(&pr_reviewers2.PrReviewerOut{}, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				PullRequestID:   prID,
				PullRequestName: req.PullRequestName,
				AuthorID:        authorID,
				Status:          "OPEN",
				AssignedReviewers: []uuid.UUID{
					reviewerID2,
					reviewerID1,
				},
				CreatedAt: createdPR.CreatedAt,
				MergedAt:  createdPR.MergedAt,
			},
		},
		{
			name: "pull request already exists",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(createdPR, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrPullRequestExists,
		},
		{
			name: "error checking pull request existence",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
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
			name: "author not found",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(nil, repository.ErrUserNotFound)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrAuthorPrNotFound,
		},
		{
			name: "error getting author",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrGetUser,
		},
		{
			name: "no active team members found",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(nil, repository.ErrUserNotFound)

				mockPRStatuses.EXPECT().
					SavePRStatus(gomock.Any(), gomock.Any()).
					Return(prStatusOut, nil)

				mockPullRequests.EXPECT().
					SavePullRequest(gomock.Any(), gomock.Any()).
					Return(createdPR, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				PullRequestID:     prID,
				PullRequestName:   req.PullRequestName,
				AuthorID:          authorID,
				Status:            "OPEN",
				AssignedReviewers: []uuid.UUID{},
				CreatedAt:         createdPR.CreatedAt,
				MergedAt:          createdPR.MergedAt,
			},
		},
		{
			name: "error getting team members",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrGetUsers,
		},
		{
			name: "error saving PR status",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockPRStatuses.EXPECT().
					SavePRStatus(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrSetPRStatus,
		},
		{
			name: "error saving pull request",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockPRStatuses.EXPECT().
					SavePRStatus(gomock.Any(), gomock.Any()).
					Return(prStatusOut, nil)

				mockPullRequests.EXPECT().
					SavePullRequest(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrSavePullRequest,
		},
		{
			name: "error assigning reviewer",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockPRStatuses.EXPECT().
					SavePRStatus(gomock.Any(), gomock.Any()).
					Return(prStatusOut, nil)

				mockPullRequests.EXPECT().
					SavePullRequest(gomock.Any(), gomock.Any()).
					Return(createdPR, nil)

				mockRandomizer.EXPECT().
					Shuffle(gomock.Any(), gomock.Any()).
					DoAndReturn(func(n int, swap func(i, j int)) {
					})

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrAssignReviewer,
		},
		{
			name: "successful create with fewer reviewers than max",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				fewTeamMembers := []users2.UserOut{
					{
						ID:       reviewerID1,
						Name:     "reviewer1",
						IsActive: true,
						TeamID:   teamID,
					},
					{
						ID:       authorID,
						Name:     "author",
						IsActive: true,
						TeamID:   teamID,
					},
				}
				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&fewTeamMembers, nil)

				mockPRStatuses.EXPECT().
					SavePRStatus(gomock.Any(), gomock.Any()).
					Return(prStatusOut, nil)

				mockPullRequests.EXPECT().
					SavePullRequest(gomock.Any(), gomock.Any()).
					Return(createdPR, nil)

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.Any()).
					Times(1).Return(&pr_reviewers2.PrReviewerOut{}, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				PullRequestID:     prID,
				PullRequestName:   req.PullRequestName,
				AuthorID:          authorID,
				Status:            "OPEN",
				AssignedReviewers: []uuid.UUID{reviewerID1},
				CreatedAt:         createdPR.CreatedAt,
				MergedAt:          createdPR.MergedAt,
			},
		},
		{
			name: "successful create with no available reviewers",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(nil, repository.ErrPullRequestNotFound)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				onlyAuthorTeam := []users2.UserOut{
					{
						ID:       authorID,
						Name:     "author",
						IsActive: true,
						TeamID:   teamID,
					},
				}
				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&onlyAuthorTeam, nil)

				mockPRStatuses.EXPECT().
					SavePRStatus(gomock.Any(), gomock.Any()).
					Return(prStatusOut, nil)

				mockPullRequests.EXPECT().
					SavePullRequest(gomock.Any(), gomock.Any()).
					Return(createdPR, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				PullRequestID:     prID,
				PullRequestName:   req.PullRequestName,
				AuthorID:          authorID,
				Status:            "OPEN",
				AssignedReviewers: []uuid.UUID{},
				CreatedAt:         createdPR.CreatedAt,
				MergedAt:          createdPR.MergedAt,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoPullRequests := pull_requests.NewMockRepositoryPullRequests(ctrl)
			mockRepoUsers := users.NewMockRepositoryUsers(ctrl)
			mockRepoPRStatuses := pr_statuses.NewMockRepositoryPrStatuses(ctrl)
			mockRepoPRReviewers := pr_reviewers.NewMockRepositoryPrReviewers(ctrl)
			mockRandomizer := randomizer.NewMockRandomizer(ctrl)
			mockTrm := mock.NewMockManager(ctrl)

			tt.setupMock(
				mockRepoPullRequests,
				mockRepoUsers,
				mockRepoPRStatuses,
				mockRepoPRReviewers,
				mockRandomizer,
				mockTrm,
			)

			u := NewUsecase(
				mockRepoUsers,
				mockRepoPullRequests,
				mockRepoPRReviewers,
				mockRepoPRStatuses,
				mockRandomizer,
				cntReviewers,
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
