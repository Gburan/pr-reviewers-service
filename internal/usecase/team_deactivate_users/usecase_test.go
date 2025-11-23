package team_deactivate_users

import (
	"context"
	"errors"
	"testing"
	"time"

	"pr-reviewers-service/internal/infrastructure/repository"
	pr_reviewers2 "pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	pr_statuses2 "pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	pull_requests2 "pr-reviewers-service/internal/infrastructure/repository/pull_requests"
	teams2 "pr-reviewers-service/internal/infrastructure/repository/teams"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	usecase2 "pr-reviewers-service/internal/usecase"
	randomizer2 "pr-reviewers-service/internal/usecase/contract/randomizer/mocks"
	pr_reviewers "pr-reviewers-service/internal/usecase/contract/repository/pr_reviewers/mocks"
	pr_statuses "pr-reviewers-service/internal/usecase/contract/repository/pr_statuses/mocks"
	pull_requests "pr-reviewers-service/internal/usecase/contract/repository/pull_requests/mocks"
	teams "pr-reviewers-service/internal/usecase/contract/repository/teams/mocks"
	users "pr-reviewers-service/internal/usecase/contract/repository/users/mocks"

	"github.com/avito-tech/go-transaction-manager/trm/v2/drivers/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamDeactivateUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamID := uuid.New()
	teamName := "test-team"
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()
	pr1ID := uuid.New()
	pr2ID := uuid.New()
	statusID := uuid.New()

	team := &teams2.TeamOut{
		ID:   teamID,
		Name: teamName,
	}

	activeUsers := []users2.UserOut{
		{
			ID:       user1ID,
			Name:     "user1",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       user2ID,
			Name:     "user2",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       user3ID,
			Name:     "user3",
			IsActive: true,
			TeamID:   teamID,
		},
	}

	inactiveUser := users2.UserOut{
		ID:       user1ID,
		Name:     "user1",
		IsActive: false,
		TeamID:   teamID,
	}

	prReviewers := []pr_reviewers2.PrReviewerOut{
		{
			ID:         uuid.New(),
			PRID:       pr1ID,
			ReviewerID: user1ID,
		},
		{
			ID:         uuid.New(),
			PRID:       pr1ID,
			ReviewerID: user2ID,
		},
		{
			ID:         uuid.New(),
			PRID:       pr2ID,
			ReviewerID: user1ID,
		},
	}

	openPRs := []pull_requests2.PullRequestOut{
		{
			ID:        pr1ID,
			Name:      "PR 1",
			AuthorID:  user3ID,
			StatusID:  statusID,
			CreatedAt: time.Now(),
		},
		{
			ID:        pr2ID,
			Name:      "PR 2",
			AuthorID:  user2ID,
			StatusID:  statusID,
			CreatedAt: time.Now(),
		},
	}

	closedPR := pull_requests2.PullRequestOut{
		ID:        uuid.New(),
		Name:      "Closed PR",
		AuthorID:  user3ID,
		StatusID:  uuid.New(),
		CreatedAt: time.Now(),
	}

	openStatus := &pr_statuses2.PRStatusOut{
		ID:     statusID,
		Status: usecase2.OpenStatusValue,
	}

	closedStatus := &pr_statuses2.PRStatusOut{
		ID:     uuid.New(),
		Status: "CLOSED",
	}

	updatedTeamMembers := []users2.UserOut{
		{
			ID:       user1ID,
			Name:     "user1",
			IsActive: false,
			TeamID:   teamID,
		},
		{
			ID:       user2ID,
			Name:     "user2",
			IsActive: false,
			TeamID:   teamID,
		},
		{
			ID:       user3ID,
			Name:     "user3",
			IsActive: true,
			TeamID:   teamID,
		},
	}

	tests := []struct {
		name      string
		req       In
		setupMock func(
			mockTeams *teams.MockRepositoryTeams,
			mockUsers *users.MockRepositoryUsers,
			mockPullRequests *pull_requests.MockRepositoryPullRequests,
			mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
			mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
			mockRandomizer *randomizer2.MockRandomizer,
			mockTrm *mock.MockManager,
		)
		expected      *Out
		expectedError error
	}{
		{
			name: "successful deactivation with PR reassignment",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID, user2ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID, user2ID}).
					Return(&[]users2.UserOut{activeUsers[0], activeUsers[1]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID, user2ID}).
					Return(&prReviewers, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID, pr2ID}).
					Return(&openPRs, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID, statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser, inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0], prReviewers[1]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
					Return(&activeUsers[2], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&[]users2.UserOut{activeUsers[2]}, nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user1ID).
					Return(nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user2ID).
					Return(nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr2ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[2]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr2ID).
					Return(&openPRs[1], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user2ID).
					Return(&activeUsers[1], nil)

				user4ID := uuid.MustParse("4a9a18c2-7c90-42a5-b839-7c052fbae8e0")
				activeUser4 := users2.UserOut{
					ID:       user4ID,
					Name:     "user4",
					IsActive: true,
					TeamID:   teamID,
				}

				availableForPR2 := []users2.UserOut{activeUsers[2], activeUser4}
				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&availableForPR2, nil)

				mockRandomizer.EXPECT().
					Shuffle(2, gomock.Any()).
					DoAndReturn(func(n int, swap func(i, j int)) {
						if n >= 2 {
							swap(0, 1)
						}
					})

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr2ID, user1ID).
					Return(nil)

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.AssignableToTypeOf(pr_reviewers2.PrReviewerIn{})).
					DoAndReturn(func(ctx context.Context, in pr_reviewers2.PrReviewerIn) (*pr_reviewers2.PrReviewerOut, error) {
						assert.Equal(t, pr2ID, in.PrID)
						assert.True(t, in.ReviewerID == user3ID || in.ReviewerID == user4ID)
						return &pr_reviewers2.PrReviewerOut{
							ID:         uuid.New(),
							PRID:       in.PrID,
							ReviewerID: in.ReviewerID,
						}, nil
					})

				updatedTeamMembersWithUser4 := []users2.UserOut{
					{ID: user1ID, Name: "user1", IsActive: false, TeamID: teamID},
					{ID: user2ID, Name: "user2", IsActive: false, TeamID: teamID},
					{ID: user3ID, Name: "user3", IsActive: true, TeamID: teamID},
					{ID: user4ID, Name: "user4", IsActive: true, TeamID: teamID},
				}
				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&updatedTeamMembersWithUser4, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				Team: Team{
					TeamName: teamName,
					Members: []TeamMember{
						{UserID: user1ID, Username: "user1", IsActive: false},
						{UserID: user2ID, Username: "user2", IsActive: false},
						{UserID: user3ID, Username: "user3", IsActive: true},
						{UserID: uuid.MustParse("4a9a18c2-7c90-42a5-b839-7c052fbae8e0"), Username: "user4", IsActive: true},
					},
				},
				AffectedPullRequests: []PullRequestShort{
					{PullRequestID: pr1ID, PullRequestName: "PR 1", AuthorID: user3ID, Status: usecase2.OpenStatusValue},
					{PullRequestID: pr2ID, PullRequestName: "PR 2", AuthorID: user2ID, Status: usecase2.OpenStatusValue},
				},
			},
		},
		{
			name: "team not found",
			req: In{
				TeamName: "non-existent-team",
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), "non-existent-team").
					Return(nil, repository.ErrTeamNotFound)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrTeamNotFound,
		},
		{
			name: "error getting team",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrGetTeam,
		},
		{
			name: "users not found",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(nil, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrUsersByIDsNotFound,
		},
		{
			name: "error getting users",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
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
			name: "user not in team",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				otherTeamID := uuid.New()
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				userFromOtherTeam := users2.UserOut{
					ID:       user1ID,
					Name:     "user1",
					IsActive: true,
					TeamID:   otherTeamID,
				}
				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{userFromOtherTeam}, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrUserNotBelongsToTeam,
		},
		{
			name: "no PRs to affect",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(nil, repository.ErrPRReviewerNotFound)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrNoUsersAssignedToPRs,
		},
		{
			name: "no users assigned to pr",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{closedPR}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{closedPR.StatusID}).
					Return(&[]pr_statuses2.PRStatusOut{*closedStatus}, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrNoPRsToAffect,
		},
		{
			name: "error getting PR reviewers",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
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
			name: "error getting pull requests",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
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
			name: "error getting PR statuses",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
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
			name: "error updating users",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrUpdateUser,
		},
		{
			name: "successful deactivation without PR reassignment - no reviewers to remove",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[1]}, nil)

				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&updatedTeamMembers, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				Team: Team{
					TeamName: teamName,
					Members: []TeamMember{
						{UserID: user1ID, Username: "user1", IsActive: false},
						{UserID: user2ID, Username: "user2", IsActive: false},
						{UserID: user3ID, Username: "user3", IsActive: true},
					},
				},
				AffectedPullRequests: []PullRequestShort{},
			},
		},
		{
			name: "error getting PR for reassignment",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
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
			name: "error getting author for reassignment",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
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
			name: "error getting team members for reassignment",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
					Return(&activeUsers[2], nil)

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
			name: "error removing reviewer",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
					Return(&activeUsers[2], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&[]users2.UserOut{activeUsers[2]}, nil)

				mockRandomizer.EXPECT().Shuffle(gomock.Any(), gomock.Any()).AnyTimes()

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user1ID).
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
			name: "successful deactivation with no available reviewers - just remove old ones",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
					Return(&activeUsers[2], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&[]users2.UserOut{activeUsers[2]}, nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user1ID).
					Return(nil)

				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&updatedTeamMembers, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				Team: Team{
					TeamName: teamName,
					Members: []TeamMember{
						{UserID: user1ID, Username: "user1", IsActive: false},
						{UserID: user2ID, Username: "user2", IsActive: false},
						{UserID: user3ID, Username: "user3", IsActive: true},
					},
				},
				AffectedPullRequests: []PullRequestShort{
					{PullRequestID: pr1ID, PullRequestName: "PR 1", AuthorID: user3ID, Status: usecase2.OpenStatusValue},
				},
			},
		},
		{
			name: "error getting updated team members",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]users2.UserOut{activeUsers[0]}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
					Return(&activeUsers[2], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&[]users2.UserOut{activeUsers[2]}, nil)

				mockRandomizer.EXPECT().Shuffle(gomock.Any(), gomock.Any()).AnyTimes()

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user1ID).
					Return(nil)

				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
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
			name: "duplicate user IDs in request",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID, user1ID, user2ID, user1ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, ids []uuid.UUID) (*[]users2.UserOut, error) {
						uniqueIDs := make(map[uuid.UUID]bool)
						for _, id := range ids {
							uniqueIDs[id] = true
						}
						assert.Equal(t, 2, len(uniqueIDs), "Should receive unique user IDs")

						return &[]users2.UserOut{activeUsers[0], activeUsers[1]}, nil
					})

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, ids []uuid.UUID) (*[]pr_reviewers2.PrReviewerOut, error) {
						uniqueIDs := make(map[uuid.UUID]bool)
						for _, id := range ids {
							uniqueIDs[id] = true
						}
						assert.Equal(t, 2, len(uniqueIDs), "Should receive unique user IDs")
						return &[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil
					})

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), gomock.Any()).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), gomock.Any()).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Len(2)).
					Return(&[]users2.UserOut{inactiveUser, inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), gomock.Any()).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), gomock.Any()).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), gomock.Any()).
					Return(&activeUsers[2], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{activeUsers[2]}, nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), gomock.Any()).
					Return(&updatedTeamMembers, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				Team: Team{
					TeamName: teamName,
					Members: []TeamMember{
						{UserID: user1ID, Username: "user1", IsActive: false},
						{UserID: user2ID, Username: "user2", IsActive: false},
						{UserID: user3ID, Username: "user3", IsActive: true},
					},
				},
				AffectedPullRequests: []PullRequestShort{
					{PullRequestID: pr1ID, PullRequestName: "PR 1", AuthorID: user3ID, Status: usecase2.OpenStatusValue},
				},
			},
		},
		{
			name: "deactivation of already inactive users",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID, user2ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				inactiveUser1 := users2.UserOut{
					ID:       user1ID,
					Name:     "user1",
					IsActive: false,
					TeamID:   teamID,
				}
				activeUser2 := users2.UserOut{
					ID:       user2ID,
					Name:     "user2",
					IsActive: true,
					TeamID:   teamID,
				}
				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID, user2ID}).
					Return(&[]users2.UserOut{inactiveUser1, activeUser2}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID, user2ID}).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[1]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Len(1)).
					DoAndReturn(func(ctx context.Context, users []users2.UserIn) (*[]users2.UserOut, error) {
						assert.Len(t, users, 1)
						assert.Equal(t, user2ID, users[0].ID)
						assert.False(t, users[0].IsActive)

						return &[]users2.UserOut{
							{
								ID:       user2ID,
								Name:     "user2",
								IsActive: false,
								TeamID:   teamID,
							},
						}, nil
					})

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[1]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
					Return(&activeUsers[2], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&[]users2.UserOut{activeUsers[2]}, nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user2ID).
					Return(nil)

				updatedMembers := []users2.UserOut{
					{ID: user1ID, Name: "user1", IsActive: false, TeamID: teamID},
					{ID: user2ID, Name: "user2", IsActive: false, TeamID: teamID},
					{ID: user3ID, Name: "user3", IsActive: true, TeamID: teamID},
				}
				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&updatedMembers, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				Team: Team{
					TeamName: teamName,
					Members: []TeamMember{
						{UserID: user1ID, Username: "user1", IsActive: false},
						{UserID: user2ID, Username: "user2", IsActive: false},
						{UserID: user3ID, Username: "user3", IsActive: true},
					},
				},
				AffectedPullRequests: []PullRequestShort{
					{PullRequestID: pr1ID, PullRequestName: "PR 1", AuthorID: user3ID, Status: usecase2.OpenStatusValue},
				},
			},
		},
		{
			name: "deactivate all team members",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user1ID, user2ID, user3ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user1ID, user2ID, user3ID}).
					Return(&activeUsers, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user1ID, user2ID, user3ID}).
					Return(&prReviewers, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID, pr2ID}).
					Return(&openPRs, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID, statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Len(3)).
					Return(&[]users2.UserOut{inactiveUser, inactiveUser, inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[0], prReviewers[1]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
					Return(&activeUsers[2], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&[]users2.UserOut{}, nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user1ID).
					Return(nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user2ID).
					Return(nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr2ID).
					Return(&[]pr_reviewers2.PrReviewerOut{prReviewers[2]}, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr2ID).
					Return(&openPRs[1], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user2ID).
					Return(&activeUsers[1], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&[]users2.UserOut{}, nil)

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr2ID, user1ID).
					Return(nil)

				allInactiveMembers := []users2.UserOut{
					{ID: user1ID, Name: "user1", IsActive: false, TeamID: teamID},
					{ID: user2ID, Name: "user2", IsActive: false, TeamID: teamID},
					{ID: user3ID, Name: "user3", IsActive: false, TeamID: teamID},
				}
				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&allInactiveMembers, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				Team: Team{
					TeamName: teamName,
					Members: []TeamMember{
						{UserID: user1ID, Username: "user1", IsActive: false},
						{UserID: user2ID, Username: "user2", IsActive: false},
						{UserID: user3ID, Username: "user3", IsActive: false},
					},
				},
				AffectedPullRequests: []PullRequestShort{
					{PullRequestID: pr1ID, PullRequestName: "PR 1", AuthorID: user3ID, Status: usecase2.OpenStatusValue},
					{PullRequestID: pr2ID, PullRequestName: "PR 2", AuthorID: user2ID, Status: usecase2.OpenStatusValue},
				},
			},
		},
		{
			name: "deactivation of PR author",
			req: In{
				TeamName: teamName,
				UserIDs:  []uuid.UUID{user3ID},
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockPullRequests *pull_requests.MockRepositoryPullRequests,
				mockPRReviewers *pr_reviewers.MockRepositoryPrReviewers,
				mockPRStatuses *pr_statuses.MockRepositoryPrStatuses,
				mockRandomizer *randomizer2.MockRandomizer,
				mockTrm *mock.MockManager,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), teamName).
					Return(team, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), []uuid.UUID{user3ID}).
					Return(&[]users2.UserOut{activeUsers[2]}, nil)

				authorAsReviewer := []pr_reviewers2.PrReviewerOut{
					{
						ID:         uuid.New(),
						PRID:       pr1ID,
						ReviewerID: user3ID,
					},
				}
				mockPRReviewers.EXPECT().
					GetPRReviewersByReviewerIDs(gomock.Any(), []uuid.UUID{user3ID}).
					Return(&authorAsReviewer, nil)

				mockPullRequests.EXPECT().
					GetPullRequestsByPrIDs(gomock.Any(), []uuid.UUID{pr1ID}).
					Return(&[]pull_requests2.PullRequestOut{openPRs[0]}, nil)

				mockPRStatuses.EXPECT().
					GetPRStatusesByIDs(gomock.Any(), []uuid.UUID{statusID}).
					Return(&[]pr_statuses2.PRStatusOut{*openStatus}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{inactiveUser}, nil)

				mockPRReviewers.EXPECT().
					GetPRReviewersByPRID(gomock.Any(), pr1ID).
					Return(&authorAsReviewer, nil)

				mockPullRequests.EXPECT().
					GetPullRequestByID(gomock.Any(), pr1ID).
					Return(&openPRs[0], nil)

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), user3ID).
					Return(&activeUsers[2], nil)

				mockUsers.EXPECT().
					GetActiveUsersByTeamID(gomock.Any(), teamID).
					Return(&[]users2.UserOut{activeUsers[0], activeUsers[1]}, nil)

				mockRandomizer.EXPECT().
					Shuffle(2, gomock.Any()).
					DoAndReturn(func(n int, swap func(i, j int)) {
						if n >= 2 {
							swap(0, 1)
						}
					})

				mockPRReviewers.EXPECT().
					DeletePRReviewerByPRAndReviewer(gomock.Any(), pr1ID, user3ID).
					Return(nil)

				mockPRReviewers.EXPECT().
					SavePRReviewer(gomock.Any(), gomock.AssignableToTypeOf(pr_reviewers2.PrReviewerIn{})).
					DoAndReturn(func(ctx context.Context, in pr_reviewers2.PrReviewerIn) (*pr_reviewers2.PrReviewerOut, error) {
						assert.Equal(t, pr1ID, in.PrID)
						assert.True(t, in.ReviewerID == user1ID || in.ReviewerID == user2ID)
						return &pr_reviewers2.PrReviewerOut{
							ID:         uuid.New(),
							PRID:       in.PrID,
							ReviewerID: in.ReviewerID,
						}, nil
					})

				updatedMembers := []users2.UserOut{
					{ID: user1ID, Name: "user1", IsActive: true, TeamID: teamID},
					{ID: user2ID, Name: "user2", IsActive: true, TeamID: teamID},
					{ID: user3ID, Name: "user3", IsActive: false, TeamID: teamID},
				}
				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&updatedMembers, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				Team: Team{
					TeamName: teamName,
					Members: []TeamMember{
						{UserID: user1ID, Username: "user1", IsActive: true},
						{UserID: user2ID, Username: "user2", IsActive: true},
						{UserID: user3ID, Username: "user3", IsActive: false},
					},
				},
				AffectedPullRequests: []PullRequestShort{
					{PullRequestID: pr1ID, PullRequestName: "PR 1", AuthorID: user3ID, Status: usecase2.OpenStatusValue},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoTeams := teams.NewMockRepositoryTeams(ctrl)
			mockRepoUsers := users.NewMockRepositoryUsers(ctrl)
			mockRepoPullRequests := pull_requests.NewMockRepositoryPullRequests(ctrl)
			mockRepoPRReviewers := pr_reviewers.NewMockRepositoryPrReviewers(ctrl)
			mockRepoPRStatuses := pr_statuses.NewMockRepositoryPrStatuses(ctrl)
			mockRandomizer := randomizer2.NewMockRandomizer(ctrl)
			mockTrm := mock.NewMockManager(ctrl)

			tt.setupMock(
				mockRepoTeams,
				mockRepoUsers,
				mockRepoPullRequests,
				mockRepoPRReviewers,
				mockRepoPRStatuses,
				mockRandomizer,
				mockTrm,
			)

			u := NewUsecase(
				mockRepoTeams,
				mockRepoUsers,
				mockRepoPullRequests,
				mockRepoPRReviewers,
				mockRepoPRStatuses,
				mockRandomizer,
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
				assert.Equal(t, tt.expected.Team.TeamName, result.Team.TeamName)
				assert.Equal(t, len(tt.expected.Team.Members), len(result.Team.Members))
				assert.Equal(t, len(tt.expected.AffectedPullRequests), len(result.AffectedPullRequests))

				for i, expectedMember := range tt.expected.Team.Members {
					assert.Equal(t, expectedMember.UserID, result.Team.Members[i].UserID)
					assert.Equal(t, expectedMember.Username, result.Team.Members[i].Username)
					assert.Equal(t, expectedMember.IsActive, result.Team.Members[i].IsActive)
				}

				for i, expectedPR := range tt.expected.AffectedPullRequests {
					assert.Equal(t, expectedPR.PullRequestID, result.AffectedPullRequests[i].PullRequestID)
					assert.Equal(t, expectedPR.PullRequestName, result.AffectedPullRequests[i].PullRequestName)
					assert.Equal(t, expectedPR.AuthorID, result.AffectedPullRequests[i].AuthorID)
					assert.Equal(t, expectedPR.Status, result.AffectedPullRequests[i].Status)
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
