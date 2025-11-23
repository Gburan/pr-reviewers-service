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

func (s *ApiTest) TestGetUserReviewPRs() {
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
		setup         func(ctx context.Context, repos *TestRepos) (userID uuid.UUID, expectedPRs []uuid.UUID)
		checkResponse func(t *testing.T, userReviewResp handler.GetUserReviewPRsResponse, expectedPRs []uuid.UUID)
	}{
		{
			name: "successful get user review PRs with multiple assignments",
			setup: func(ctx context.Context, repos *TestRepos) (userID uuid.UUID, expectedPRs []uuid.UUID) {
				teamID := uuid.New()
				user1ID := uuid.New()
				userID = uuid.New()
				statusID := uuid.New()
				pr1ID := uuid.New()
				pr2ID := uuid.New()
				pr3ID := uuid.New()
				now := time.Now()

				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team For Review PRs 1",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     user1ID,
						Name:   "Author 1",
						TeamID: teamID,
					},
					{
						ID:     userID,
						Name:   "Reviewer 1",
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
					Name:      "PR 1 For Review",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        pr2ID,
					Name:      "PR 2 For Review",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.PR.SavePullRequest(ctx, pull_requests.PullRequestIn{
					ID:        pr3ID,
					Name:      "PR 3 For Review",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr1ID,
					ReviewerID: userID,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr2ID,
					ReviewerID: userID,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr3ID,
					ReviewerID: user1ID,
				})
				assert.NoError(s.T(), err)

				expectedPRs = []uuid.UUID{pr1ID, pr2ID}
				return userID, expectedPRs
			},
			checkResponse: func(t *testing.T, userReviewResp handler.GetUserReviewPRsResponse, expectedPRs []uuid.UUID) {
				assert.NotNil(t, userReviewResp)
				assert.Len(t, userReviewResp.PullRequests, len(expectedPRs))

				expectedPRsMap := make(map[uuid.UUID]bool)
				for _, prID := range expectedPRs {
					expectedPRsMap[prID] = true
				}

				for _, pr := range userReviewResp.PullRequests {
					assert.True(t, expectedPRsMap[pr.PullRequestId],
						"Unexpected PR %s in response", pr.PullRequestId)
					assert.NotEmpty(t, pr.PullRequestName)
					assert.NotEqual(t, uuid.Nil, pr.AuthorId)
					assert.Equal(t, "open", string(pr.Status))
				}
			},
		},
		{
			name: "user with no review assignments returns empty list",
			setup: func(ctx context.Context, repos *TestRepos) (userID uuid.UUID, expectedPRs []uuid.UUID) {
				teamID := uuid.New()
				user1ID := uuid.New()
				userID = uuid.New()
				statusID := uuid.New()
				pr1ID := uuid.New()
				now := time.Now()

				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team For Review PRs 2",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     user1ID,
						Name:   "Author 2",
						TeamID: teamID,
					},
					{
						ID:     userID,
						Name:   "Reviewer Without Assignments",
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
					Name:      "PR For No Assignments",
					AuthorID:  user1ID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       pr1ID,
					ReviewerID: user1ID,
				})
				assert.NoError(s.T(), err)

				expectedPRs = []uuid.UUID{}
				return userID, expectedPRs
			},
			checkResponse: func(t *testing.T, userReviewResp handler.GetUserReviewPRsResponse, expectedPRs []uuid.UUID) {
				assert.NotNil(t, userReviewResp)
				assert.Empty(t, userReviewResp.PullRequests)
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

			userID, expectedPRs := tt.setup(ctx, repos)

			userReviewResp, status, _, err := getUserReviewPRs(token, userID.String())

			assert.NoError(t, err)
			assert.True(t, status == http.StatusOK || status == http.StatusNotFound,
				"Expected 200 or 404, got %d", status)

			if tt.checkResponse != nil {
				tt.checkResponse(t, userReviewResp, expectedPRs)
			}
		})
	}
}
