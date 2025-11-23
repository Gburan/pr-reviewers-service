package pull_request_reassign

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
	randomizer2 "pr-reviewers-service/internal/usecase/contract/randomizer/mocks"
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

func TestPullRequestReassign(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prID := uuid.New()
	authorID := uuid.New()
	teamID := uuid.New()
	statusID := uuid.New()
	oldUserID := uuid.New()
	newUserID := uuid.New()
	reviewerID1 := uuid.New()
	reviewerID2 := uuid.New()
	newuserID2 := uuid.New()

	req := In{
		PullRequestID: prID,
		OldUserId:     oldUserID,
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
	author := &users2.UserOut{
		ID:       authorID,
		Name:     "author",
		IsActive: true,
		TeamID:   teamID,
	}
	currentReviewers := []pr_reviewers2.PrReviewerOut{
		{
			ID:         uuid.New(),
			PRID:       prID,
			ReviewerID: oldUserID,
		},
		{
			ID:         uuid.New(),
			PRID:       prID,
			ReviewerID: reviewerID1,
		},
	}
	teamMembers := []users2.UserOut{
		{
			ID:       authorID,
			Name:     "author",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       oldUserID,
			Name:     "old_reviewer",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       reviewerID1,
			Name:     "reviewer1",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       newUserID,
			Name:     "new_reviewer",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       reviewerID2,
			Name:     "reviewer2",
			IsActive: false,
			TeamID:   teamID,
		},
	}
	updatedReviewers := []pr_reviewers2.PrReviewerOut{
		{
			ID:         uuid.New(),
			PRID:       prID,
			ReviewerID: reviewerID1,
		},
		{
			ID:         uuid.New(),
			PRID:       prID,
			ReviewerID: newUserID,
		},
	}

	tests := []struct {
		name      string
		req       In
		setupMock func(
			mockPullRequests *pull_requests.MockRepositoryPullRequests,
			mockUsers *users.MockRepositoryUsers,
			mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
			mockRandomizer *randomizer2.MockRandomizer,
			mockTrm *mock.MockManager,
		)
		expected      *Out
		expectedError error
	}{
		{
			name: "successful reassign pull request reviewer",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockRandomizer.EXPECT().Shuffle(gomock.Any(), gomock.Any()).AnyTimes()

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), prID, oldUserID).
					Return(nil)

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.Any()).
					Return(&pr_reviewers2.PrReviewerOut{}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&updatedReviewers, nil)

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
				Status:          "OPEN",
				AssignedReviewers: []uuid.UUID{
					reviewerID1,
					newUserID,
				},
				CreatedAt:  existingPR.CreatedAt,
				MergedAt:   existingPR.MergedAt,
				ReplacedBy: newUserID,
			},
		},
		{
			name: "successful reassign pull request reviewer - more than one available reviewer",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				teamCopy := make([]users2.UserOut, len(teamMembers))
				copy(teamCopy, teamMembers)
				teamCopy = append(teamCopy, users2.UserOut{
					ID:       newuserID2,
					Name:     "new_reviewer2",
					IsActive: true,
					TeamID:   teamID,
				})
				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamCopy, nil)

				mockRandomizer.EXPECT().
					Shuffle(2, gomock.Any()).
					DoAndReturn(func(n int, swap func(i, j int)) {
						swap(0, 1)
					})

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), prID, oldUserID).
					Return(nil)

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.Any()).
					Return(&pr_reviewers2.PrReviewerOut{}, nil)

				updatedReviewers2 := []pr_reviewers2.PrReviewerOut{
					{
						ID:         uuid.New(),
						PRID:       prID,
						ReviewerID: reviewerID1,
					},
					{
						ID:         uuid.New(),
						PRID:       prID,
						ReviewerID: newuserID2,
					},
				}
				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&updatedReviewers2, nil)

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
				Status:          "OPEN",
				AssignedReviewers: []uuid.UUID{
					reviewerID1,
					newuserID2,
				},
				CreatedAt:  existingPR.CreatedAt,
				MergedAt:   existingPR.MergedAt,
				ReplacedBy: newuserID2,
			},
		},
		{
			name: "pull request not found",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
			name: "pull request already merged",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(mergedStatus, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrPullRequestAlreadyMerged,
		},
		{
			name: "error getting current reviewers",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
			name: "old reviewer not found in current reviewers",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), prID).
					Return(existingPR, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusByID(gomock.Any(), pr_statuses2.PRStatusIn{ID: statusID}).
					Return(currentStatus, nil)

				reviewersWithoutOld := []pr_reviewers2.PrReviewerOut{
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
				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&reviewersWithoutOld, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrReviewerNotFound,
		},
		{
			name: "author not found",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

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
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

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
			name: "error getting team members",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

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
			name: "no available reviewers",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				limitedTeam := []users2.UserOut{
					{
						ID:       authorID,
						Name:     "author",
						IsActive: true,
						TeamID:   teamID,
					},
					{
						ID:       oldUserID,
						Name:     "old_reviewer",
						IsActive: true,
						TeamID:   teamID,
					},
					{
						ID:       reviewerID1,
						Name:     "reviewer1",
						IsActive: true,
						TeamID:   teamID,
					},
				}
				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&limitedTeam, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrNoAvailableReviewers,
		},
		{
			name: "error removing old reviewer",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockRandomizer.EXPECT().Shuffle(gomock.Any(), gomock.Any()).AnyTimes()

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), prID, oldUserID).
					Return(errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrRemoveReviewer,
		},
		{
			name: "error assigning new reviewer",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockRandomizer.EXPECT().Shuffle(gomock.Any(), gomock.Any()).AnyTimes()

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), prID, oldUserID).
					Return(nil)

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
			name: "error getting updated reviewers",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockRandomizer.EXPECT().Shuffle(gomock.Any(), gomock.Any()).AnyTimes()

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), prID, oldUserID).
					Return(nil)

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.Any()).
					Return(&pr_reviewers2.PrReviewerOut{}, nil)

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
			name: "successful reassign with no reviewers found after update",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&teamMembers, nil)

				mockRandomizer.EXPECT().Shuffle(gomock.Any(), gomock.Any()).AnyTimes()

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), prID, oldUserID).
					Return(nil)

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.Any()).
					Return(&pr_reviewers2.PrReviewerOut{}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(nil, repository.ErrPRReviewerNotFound)

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
				Status:            "OPEN",
				AssignedReviewers: []uuid.UUID{},
				CreatedAt:         existingPR.CreatedAt,
				MergedAt:          existingPR.MergedAt,
				ReplacedBy:        newUserID,
			},
		},
		{
			name: "successful reassign with single available reviewer - no shuffle",
			req:  req,
			setupMock: func(
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockUsers *users.MockRepositoryUsers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockRandomizer *randomizer2.MockRandomizer,
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
					Return(&currentReviewers, nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), authorID).
					Return(author, nil)

				singleAvailableTeam := []users2.UserOut{
					{
						ID:       authorID,
						Name:     "author",
						IsActive: true,
						TeamID:   teamID,
					},
					{
						ID:       oldUserID,
						Name:     "old_reviewer",
						IsActive: true,
						TeamID:   teamID,
					},
					{
						ID:       reviewerID1,
						Name:     "reviewer1",
						IsActive: true,
						TeamID:   teamID,
					},
					{
						ID:       newUserID,
						Name:     "new_reviewer",
						IsActive: true,
						TeamID:   teamID,
					},
				}
				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&singleAvailableTeam, nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), prID, oldUserID).
					Return(nil)

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.Any()).
					Return(&pr_reviewers2.PrReviewerOut{}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), prID).
					Return(&updatedReviewers, nil)

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
				Status:          "OPEN",
				AssignedReviewers: []uuid.UUID{
					reviewerID1,
					newUserID,
				},
				CreatedAt:  existingPR.CreatedAt,
				MergedAt:   existingPR.MergedAt,
				ReplacedBy: newUserID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoPullRequests := pull_requests.NewMockRepositoryPullRequests(ctrl)
			mockRepoUsers := users.NewMockRepositoryUsers(ctrl)
			mockRepoPRStatuses := pr_statuses.NewMockRepositoryPrStatuses(ctrl)
			mockRepoPRReviewers := pr_reviewers.NewMockRepositoryPrReviewers(ctrl)
			mockRandomizer := randomizer2.NewMockRandomizer(ctrl)
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
				assert.Equal(t, tt.expected.ReplacedBy, result.ReplacedBy)

				for i, expectedReviewer := range tt.expected.AssignedReviewers {
					assert.Equal(t, expectedReviewer, result.AssignedReviewers[i])
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
