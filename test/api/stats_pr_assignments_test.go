package api

import (
	"context"
	"net/http"
	"testing"
	"time"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	nower2 "pr-reviewers-service/internal/infrastructure/nower"
	"pr-reviewers-service/internal/infrastructure/repository/pr_reviewers"
	"pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	"pr-reviewers-service/internal/infrastructure/repository/pull_requests"
	"pr-reviewers-service/internal/infrastructure/repository/teams"
	"pr-reviewers-service/internal/infrastructure/repository/users"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *ApiTest) TestGetReviewersStats() {
	loginResp, status, _, err := dummyLogin()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)
	token := loginResp.Token

	type TestRepos struct {
		Team     *teams.Repository
		User     *users.Repository
		Status   *pr_statuses.Repository
		PR       *pull_requests.Repository
		Reviewer *pr_reviewers.Repository
	}

	tests := []struct {
		name          string
		setup         func(ctx context.Context, repos *TestRepos) (expectedStats handler.ReviewersStatsResponse)
		checkResponse func(t *testing.T, statsResp handler.ReviewersStatsResponse, expectedStats handler.ReviewersStatsResponse)
	}{
		{
			name: "successful get reviewers stats with multiple PRs",
			setup: func(ctx context.Context, repos *TestRepos) (expectedStats handler.ReviewersStatsResponse) {
				teamID := uuid.New()
				user1ID := uuid.New()
				user2ID := uuid.New()
				user3ID := uuid.New()
				statusID := uuid.New()
				pr1ID := uuid.New()
				pr2ID := uuid.New()
				pr3ID := uuid.New()
				now := time.Now()

				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Development Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     user1ID,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:     user2ID,
						Name:   "Reviewer 1",
						TeamID: teamID,
					},
					{
						ID:     user3ID,
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
					ID:        pr1ID,
					Name:      "PR 1",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        pr2ID,
					Name:      "PR 2",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        pr3ID,
					Name:      "PR 3",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr1ID,
					ReviewerID: user2ID,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr1ID,
					ReviewerID: user3ID,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr2ID,
					ReviewerID: user2ID,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr3ID,
					ReviewerID: user3ID,
				})
				assert.NoError(s.T(), err)

				expectedStats = handler.ReviewersStatsResponse{
					Reviewers: []handler.ReviewerAssignmentCount{
						{
							AssignmentCount: 2,
							ReviewerId:      user2ID,
						},
						{
							AssignmentCount: 2,
							ReviewerId:      user3ID,
						},
					},
				}
				return expectedStats
			},
			checkResponse: func(t *testing.T, statsResp handler.ReviewersStatsResponse, expectedStats handler.ReviewersStatsResponse) {
				assert.NotNil(t, statsResp)

				expectedMap := make(map[uuid.UUID]int)
				for _, stat := range expectedStats.Reviewers {
					expectedMap[stat.ReviewerId] = stat.AssignmentCount
				}

				for _, stat := range statsResp.Reviewers {
					if expectedCount, exists := expectedMap[stat.ReviewerId]; exists {
						assert.Equal(t, expectedCount, stat.AssignmentCount,
							"Reviewer %s should have %d PRs, but has %d",
							stat.ReviewerId, expectedCount, stat.AssignmentCount)
					}
				}
			},
		},
		{
			name: "successful get reviewers stats with single reviewer",
			setup: func(ctx context.Context, repos *TestRepos) (expectedStats handler.ReviewersStatsResponse) {
				teamID := uuid.New()
				user1ID := uuid.New()
				user2ID := uuid.New()
				statusID := uuid.New()
				pr1ID := uuid.New()
				pr2ID := uuid.New()
				now := time.Now()

				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Solo Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     user1ID,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:     user2ID,
						Name:   "Solo Reviewer",
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
					ID:        pr1ID,
					Name:      "PR 1",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        pr2ID,
					Name:      "PR 2",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr1ID,
					ReviewerID: user2ID,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr2ID,
					ReviewerID: user2ID,
				})
				assert.NoError(s.T(), err)

				expectedStats = handler.ReviewersStatsResponse{
					Reviewers: []handler.ReviewerAssignmentCount{
						{
							AssignmentCount: 2,
							ReviewerId:      user2ID,
						},
					},
				}
				return expectedStats
			},
			checkResponse: func(t *testing.T, statsResp handler.ReviewersStatsResponse, expectedStats handler.ReviewersStatsResponse) {
				assert.NotNil(t, statsResp)

				found := false
				for _, stat := range statsResp.Reviewers {
					if stat.ReviewerId == expectedStats.Reviewers[0].ReviewerId {
						assert.Equal(t, expectedStats.Reviewers[0].AssignmentCount, stat.AssignmentCount)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected reviewer %s not found in response", expectedStats.Reviewers[0].ReviewerId)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			repos := &TestRepos{
				Team:     teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:     users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status:   pr_statuses.NewRepository(suite2.GlobalPool),
				PR:       pull_requests.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Reviewer: pr_reviewers.NewRepository(suite2.GlobalPool),
			}

			expectedStats := tt.setup(ctx, repos)

			statsResp, status, _, err := getReviewersStats(token)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, status)

			if tt.checkResponse != nil {
				tt.checkResponse(t, statsResp, expectedStats)
			}
		})
	}
}
