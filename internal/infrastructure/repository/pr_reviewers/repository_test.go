package pr_reviewers

import (
	"context"
	"testing"
	"time"

	nower2 "pr-reviewers-service/internal/infrastructure/nower"
	"pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	"pr-reviewers-service/internal/infrastructure/repository/pull_requests"
	"pr-reviewers-service/internal/infrastructure/repository/teams"
	"pr-reviewers-service/internal/infrastructure/repository/users"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *PRReviewersTest) TestSavePRReviewer() {
	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	statusID := uuid.New()
	prID := uuid.New()
	prReviewerID := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team     *teams.Repository
		User     *users.Repository
		Status   *pr_statuses.Repository
		PR       *pull_requests.Repository
		Reviewer *Repository
	}

	tests := []struct {
		name        string
		input       PrReviewerIn
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *PrReviewerOut)
	}{
		{
			name: "successful SavePRReviewer with given UUID",
			input: PrReviewerIn{
				ID:         prReviewerID,
				PrID:       prID,
				ReviewerID: userID2,
			},
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID,
					Name:      "Test PR",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Equal(t, prReviewerID, result.ID)
				assert.Equal(t, prID, result.PRID)
				assert.Equal(t, userID2, result.ReviewerID)
			},
		},
		{
			name: "SavePRReviewer with zero UUID generates new one",
			input: PrReviewerIn{
				PrID:       prID,
				ReviewerID: userID2,
			},
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID,
					Name:      "Test PR",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PrReviewerOut) {
				assert.NotNil(t, result)
				assert.NotEqual(t, uuid.Nil, result.ID)
				assert.Equal(t, prID, result.PRID)
				assert.Equal(t, userID2, result.ReviewerID)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:     teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:     users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status:   pr_statuses.NewRepository(suite2.GlobalPool),
				PR:       pull_requests.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Reviewer: NewRepository(suite2.GlobalPool),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.Reviewer.SavePRReviewer(ctx, tt.input)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

func (s *PRReviewersTest) TestGetPRReviewersByPRID() {
	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()
	statusID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	prReviewerID1 := uuid.New()
	prReviewerID2 := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team     *teams.Repository
		User     *users.Repository
		Status   *pr_statuses.Repository
		PR       *pull_requests.Repository
		Reviewer *Repository
	}

	tests := []struct {
		name        string
		input       uuid.UUID
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]PrReviewerOut)
	}{
		{
			name:  "successful GetPRReviewersByPRID returns reviewers",
			input: prID1,
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer 1",
						TeamID: teamID,
					},
					{
						ID:     userID3,
						Name:   "Reviewer 2",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID1,
					Name:      "Test PR 1",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID2,
					Name:      "Test PR 2",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID1,
					PrID:       prID1,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID2,
					PrID:       prID1,
					ReviewerID: userID3,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				reviewers := *result
				reviewerIDs := make(map[uuid.UUID]bool)
				for _, reviewer := range reviewers {
					reviewerIDs[reviewer.ReviewerID] = true
					assert.Equal(t, prID1, reviewer.PRID)
				}

				assert.True(t, reviewerIDs[userID2])
				assert.True(t, reviewerIDs[userID3])
			},
		},
		{
			name:  "no reviewers for PR ID",
			input: uuid.New(),
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID1,
					Name:      "Test PR",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:     teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:     users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status:   pr_statuses.NewRepository(suite2.GlobalPool),
				PR:       pull_requests.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Reviewer: NewRepository(suite2.GlobalPool),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.Reviewer.GetPRReviewersByPRID(ctx, tt.input)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

func (s *PRReviewersTest) TestGetPRReviewersByReviewerID() {
	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()
	statusID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	prReviewerID1 := uuid.New()
	prReviewerID2 := uuid.New()
	prReviewerID3 := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team     *teams.Repository
		User     *users.Repository
		Status   *pr_statuses.Repository
		PR       *pull_requests.Repository
		Reviewer *Repository
	}

	tests := []struct {
		name        string
		input       uuid.UUID
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]PrReviewerOut)
	}{
		{
			name:  "successful GetPRReviewersByReviewerID returns PRs for reviewer",
			input: userID2,
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author 1",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer",
						TeamID: teamID,
					},
					{
						ID:     userID3,
						Name:   "Author 2",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID1,
					Name:      "Test PR 1",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID2,
					Name:      "Test PR 2",
					AuthorID:  userID3,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID1,
					PrID:       prID1,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID2,
					PrID:       prID2,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID3,
					PrID:       prID1,
					ReviewerID: userID3,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				reviewers := *result
				prIDs := make(map[uuid.UUID]bool)
				for _, reviewer := range reviewers {
					prIDs[reviewer.PRID] = true
					assert.Equal(t, userID2, reviewer.ReviewerID)
				}

				assert.True(t, prIDs[prID1])
				assert.True(t, prIDs[prID2])
			},
		},
		{
			name:  "no PRs for reviewer ID",
			input: uuid.New(),
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID1,
					Name:      "Test PR",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID1,
					PrID:       prID1,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:     teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:     users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status:   pr_statuses.NewRepository(suite2.GlobalPool),
				PR:       pull_requests.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Reviewer: NewRepository(suite2.GlobalPool),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.Reviewer.GetPRReviewersByReviewerID(ctx, tt.input)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

func (s *PRReviewersTest) TestGetPRReviewersByReviewerIDs() {
	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()
	userID4 := uuid.New()
	statusID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	prID3 := uuid.New()
	prReviewerID1 := uuid.New()
	prReviewerID2 := uuid.New()
	prReviewerID3 := uuid.New()
	prReviewerID4 := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team     *teams.Repository
		User     *users.Repository
		Status   *pr_statuses.Repository
		PR       *pull_requests.Repository
		Reviewer *Repository
	}

	tests := []struct {
		name        string
		input       []uuid.UUID
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]PrReviewerOut)
	}{
		{
			name:  "successful GetPRReviewersByReviewerIDs returns PRs for multiple reviewers",
			input: []uuid.UUID{userID2, userID3},
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author 1",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer 1",
						TeamID: teamID,
					},
					{
						ID:     userID3,
						Name:   "Reviewer 2",
						TeamID: teamID,
					},
					{
						ID:     userID4,
						Name:   "Reviewer 3",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID1,
					Name:      "Test PR 1",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID2,
					Name:      "Test PR 2",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID3,
					Name:      "Test PR 3",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID1,
					PrID:       prID1,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID2,
					PrID:       prID2,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID3,
					PrID:       prID2,
					ReviewerID: userID3,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID4,
					PrID:       prID3,
					ReviewerID: userID3,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         uuid.New(),
					PrID:       prID1,
					ReviewerID: userID4,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 4)

				reviewers := *result
				reviewerPRs := make(map[uuid.UUID][]uuid.UUID)
				for _, reviewer := range reviewers {
					reviewerPRs[reviewer.ReviewerID] = append(reviewerPRs[reviewer.ReviewerID], reviewer.PRID)
				}

				assert.Len(t, reviewerPRs[userID2], 2)
				assert.Contains(t, reviewerPRs[userID2], prID1)
				assert.Contains(t, reviewerPRs[userID2], prID2)

				assert.Len(t, reviewerPRs[userID3], 2)
				assert.Contains(t, reviewerPRs[userID3], prID2)
				assert.Contains(t, reviewerPRs[userID3], prID3)

				assert.NotContains(t, reviewerPRs, userID4)
			},
		},
		{
			name:  "empty reviewer IDs list returns empty result",
			input: []uuid.UUID{},
			setup: func(ctx context.Context, repos *TestRepos) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
		{
			name:  "no PRs for reviewer IDs",
			input: []uuid.UUID{uuid.New(), uuid.New()},
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID1,
					Name:      "Test PR",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID1,
					PrID:       prID1,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:     teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:     users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status:   pr_statuses.NewRepository(suite2.GlobalPool),
				PR:       pull_requests.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Reviewer: NewRepository(suite2.GlobalPool),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.Reviewer.GetPRReviewersByReviewerIDs(ctx, tt.input)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

func (s *PRReviewersTest) TestGetAllPRReviewers() {
	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()
	statusID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	prReviewerID1 := uuid.New()
	prReviewerID2 := uuid.New()
	prReviewerID3 := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team     *teams.Repository
		User     *users.Repository
		Status   *pr_statuses.Repository
		PR       *pull_requests.Repository
		Reviewer *Repository
	}

	tests := []struct {
		name        string
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]PrReviewerOut)
	}{
		{
			name: "successful GetAllPRReviewers returns all reviewers",
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author 1",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer 1",
						TeamID: teamID,
					},
					{
						ID:     userID3,
						Name:   "Reviewer 2",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID1,
					Name:      "Test PR 1",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID2,
					Name:      "Test PR 2",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID1,
					PrID:       prID1,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID2,
					PrID:       prID1,
					ReviewerID: userID3,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID3,
					PrID:       prID2,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 3)

				reviewers := *result
				reviewerCounts := make(map[uuid.UUID]int)
				prCounts := make(map[uuid.UUID]int)

				for _, reviewer := range reviewers {
					reviewerCounts[reviewer.ReviewerID]++
					prCounts[reviewer.PRID]++
				}

				assert.Equal(t, 2, reviewerCounts[userID2])
				assert.Equal(t, 1, reviewerCounts[userID3])
				assert.Equal(t, 2, prCounts[prID1])
				assert.Equal(t, 1, prCounts[prID2])
			},
		},
		{
			name: "no PR reviewers returns empty list",
			setup: func(ctx context.Context, repos *TestRepos) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PrReviewerOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:     teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:     users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status:   pr_statuses.NewRepository(suite2.GlobalPool),
				PR:       pull_requests.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Reviewer: NewRepository(suite2.GlobalPool),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.Reviewer.GetAllPRReviewers(ctx)
			tt.checkErr(t, err)
			tt.checkResult(t, result)
		})
	}
}

func (s *PRReviewersTest) TestDeletePRReviewerByPRAndReviewer() {
	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()
	statusID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	prReviewerID1 := uuid.New()
	prReviewerID2 := uuid.New()
	prReviewerID3 := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team     *teams.Repository
		User     *users.Repository
		Status   *pr_statuses.Repository
		PR       *pull_requests.Repository
		Reviewer *Repository
	}

	tests := []struct {
		name       string
		prID       uuid.UUID
		reviewerID uuid.UUID
		setup      func(ctx context.Context, repos *TestRepos)
		checkErr   assert.ErrorAssertionFunc
		verify     func(t *testing.T, ctx context.Context, repos *TestRepos)
	}{
		{
			name:       "successful DeletePRReviewerByPRAndReviewer",
			prID:       prID1,
			reviewerID: userID2,
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID1,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:     userID2,
						Name:   "Reviewer 1",
						TeamID: teamID,
					},
					{
						ID:     userID3,
						Name:   "Reviewer 2",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID1,
					Name:      "Test PR 1",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        prID2,
					Name:      "Test PR 2",
					AuthorID:  userID1,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID1,
					PrID:       prID1,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID2,
					PrID:       prID1,
					ReviewerID: userID3,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, PrReviewerIn{
					ID:         prReviewerID3,
					PrID:       prID2,
					ReviewerID: userID2,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			verify: func(t *testing.T, ctx context.Context, repos *TestRepos) {
				reviewersPR1, err := repos.Reviewer.GetPRReviewersByPRID(ctx, prID1)
				assert.NoError(t, err)
				assert.Len(t, *reviewersPR1, 1)
				assert.Equal(t, userID3, (*reviewersPR1)[0].ReviewerID)

				reviewersPR2, err := repos.Reviewer.GetPRReviewersByPRID(ctx, prID2)
				assert.NoError(t, err)
				assert.Len(t, *reviewersPR2, 1)
				assert.Equal(t, userID2, (*reviewersPR2)[0].ReviewerID)

				user2PRs, err := repos.Reviewer.GetPRReviewersByReviewerID(ctx, userID2)
				assert.NoError(t, err)
				assert.Len(t, *user2PRs, 1)
				assert.Equal(t, prID2, (*user2PRs)[0].PRID)
			},
		},
		{
			name:       "delete non-existent PR reviewer returns no error",
			prID:       uuid.New(),
			reviewerID: uuid.New(),
			setup:      func(ctx context.Context, repos *TestRepos) {},
			checkErr:   assert.NoError,
			verify:     func(t *testing.T, ctx context.Context, repos *TestRepos) {},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:     teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:     users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status:   pr_statuses.NewRepository(suite2.GlobalPool),
				PR:       pull_requests.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Reviewer: NewRepository(suite2.GlobalPool),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			err := repos.Reviewer.DeletePRReviewerByPRAndReviewer(ctx, tt.prID, tt.reviewerID)
			tt.checkErr(t, err)

			if tt.verify != nil {
				tt.verify(t, ctx, repos)
			}
		})
	}
}
