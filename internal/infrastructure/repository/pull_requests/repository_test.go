package pull_requests

import (
	"context"
	"testing"
	"time"

	nower2 "pr-reviewers-service/internal/infrastructure/nower"
	"pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	"pr-reviewers-service/internal/infrastructure/repository/teams"
	"pr-reviewers-service/internal/infrastructure/repository/users"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *PullRequestsTest) TestSavePullRequest() {
	teamID := uuid.New()
	userID := uuid.New()
	statusID := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team   *teams.Repository
		User   *users.Repository
		Status *pr_statuses.Repository
		PR     *Repository
	}

	tests := []struct {
		name        string
		input       PullRequestIn
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *PullRequestOut, expectedInput PullRequestIn)
	}{
		{
			name: "successful SavePullRequest with given UUID",
			input: PullRequestIn{
				ID:        uuid.New(),
				Name:      "Test PR",
				AuthorID:  userID,
				StatusID:  statusID,
				CreatedAt: now,
			},
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID,
						Name:   "Test User",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PullRequestOut, expectedInput PullRequestIn) {
				assert.NotNil(t, result)
				assert.Equal(t, expectedInput.ID, result.ID)
				assert.Equal(t, expectedInput.Name, result.Name)
				assert.Equal(t, expectedInput.AuthorID, result.AuthorID)
				assert.Equal(t, expectedInput.StatusID, result.StatusID)
				assert.Equal(t, expectedInput.CreatedAt, result.CreatedAt)
				assert.Equal(t, expectedInput.MergedAt, result.MergedAt)
			},
		},
		// ... остальные тестовые кейсы без изменений
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:   teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:   users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status: pr_statuses.NewRepository(suite2.GlobalPool),
				PR:     NewRepository(suite2.GlobalPool, nower2.Nower{}),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.PR.SavePullRequest(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result, tt.input)
			}
		})
	}
}

func (s *PullRequestsTest) TestGetPullRequestByID() {
	teamID := uuid.New()
	userID := uuid.New()
	statusID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team   *teams.Repository
		User   *users.Repository
		Status *pr_statuses.Repository
		PR     *Repository
	}

	tests := []struct {
		name        string
		input       uuid.UUID
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *PullRequestOut)
	}{
		{
			name:  "successful GetPullRequestByID returns PR",
			input: prID1,
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID,
						Name:   "Test User",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, PullRequestIn{
					ID:        prID1,
					Name:      "Test PR 1",
					AuthorID:  userID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, PullRequestIn{
					ID:        prID2,
					Name:      "Test PR 2",
					AuthorID:  userID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PullRequestOut) {
				assert.NotNil(t, result)
				assert.Equal(t, prID1, result.ID)
				assert.Equal(t, "Test PR 1", result.Name)
				assert.Equal(t, userID, result.AuthorID)
				assert.Equal(t, statusID, result.StatusID)
			},
		},
		// ... остальные тестовые кейсы без изменений
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:   teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:   users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status: pr_statuses.NewRepository(suite2.GlobalPool),
				PR:     NewRepository(suite2.GlobalPool, nower2.Nower{}),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.PR.GetPullRequestByID(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *PullRequestsTest) TestGetPullRequestsByPrIDs() {
	teamID := uuid.New()
	userID := uuid.New()
	statusID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	prID3 := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team   *teams.Repository
		User   *users.Repository
		Status *pr_statuses.Repository
		PR     *Repository
	}

	tests := []struct {
		name        string
		input       []uuid.UUID
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]PullRequestOut)
	}{
		{
			name:  "successful GetPullRequestsByPrIDs returns multiple PRs",
			input: []uuid.UUID{prID1, prID2},
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID,
						Name:   "Test User",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, PullRequestIn{
					ID:        prID1,
					Name:      "Test PR 1",
					AuthorID:  userID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, PullRequestIn{
					ID:        prID2,
					Name:      "Test PR 2",
					AuthorID:  userID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, PullRequestIn{
					ID:        prID3,
					Name:      "Test PR 3",
					AuthorID:  userID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PullRequestOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				prs := *result
				prMap := make(map[uuid.UUID]string)
				for _, pr := range prs {
					prMap[pr.ID] = pr.Name
				}

				assert.Equal(t, "Test PR 1", prMap[prID1])
				assert.Equal(t, "Test PR 2", prMap[prID2])
				assert.NotContains(t, prMap, prID3)
			},
		},
		// ... остальные тестовые кейсы без изменений
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:   teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:   users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status: pr_statuses.NewRepository(suite2.GlobalPool),
				PR:     NewRepository(suite2.GlobalPool, nower2.Nower{}),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.PR.GetPullRequestsByPrIDs(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *PullRequestsTest) TestMarkPullRequestMergedByID() {
	teamID := uuid.New()
	userID := uuid.New()
	statusID := uuid.New()
	prID1 := uuid.New()
	prID2 := uuid.New()
	now := time.Now()

	type TestRepos struct {
		Team   *teams.Repository
		User   *users.Repository
		Status *pr_statuses.Repository
		PR     *Repository
	}

	tests := []struct {
		name        string
		input       uuid.UUID
		setup       func(ctx context.Context, repos *TestRepos)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *PullRequestOut)
		verify      func(t *testing.T, ctx context.Context, repos *TestRepos)
	}{
		{
			name:  "successful MarkPullRequestMergedByID",
			input: prID1,
			setup: func(ctx context.Context, repos *TestRepos) {
				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID,
						Name:   "Test User",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, PullRequestIn{
					ID:        prID1,
					Name:      "Test PR 1",
					AuthorID:  userID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, PullRequestIn{
					ID:        prID2,
					Name:      "Test PR 2",
					AuthorID:  userID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PullRequestOut) {
				assert.NotNil(t, result)
				assert.Equal(t, prID1, result.ID)
				assert.False(t, result.MergedAt.IsZero())
				assert.WithinDuration(t, time.Now(), result.MergedAt, time.Second)
			},
			verify: func(t *testing.T, ctx context.Context, repos *TestRepos) {
				updatedPR, err := repos.PR.GetPullRequestByID(ctx, prID1)
				assert.NoError(t, err)
				assert.False(t, updatedPR.MergedAt.IsZero())

				otherPR, err := repos.PR.GetPullRequestByID(ctx, prID2)
				assert.NoError(t, err)
				assert.True(t, otherPR.MergedAt.IsZero())
			},
		},
		// ... остальные тестовые кейсы без изменений
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repos := &TestRepos{
				Team:   teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:   users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status: pr_statuses.NewRepository(suite2.GlobalPool),
				PR:     NewRepository(suite2.GlobalPool, nower2.Nower{}),
			}

			if tt.setup != nil {
				tt.setup(ctx, repos)
			}

			result, err := repos.PR.MarkPullRequestMergedByID(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
			if tt.verify != nil {
				tt.verify(t, ctx, repos)
			}
		})
	}
}
