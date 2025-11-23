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

func (s *ApiTest) TestReassignPullRequest() {
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
		setup         func(ctx context.Context, repos *TestRepos) (prID uuid.UUID, oldReviewerID uuid.UUID, newReviewerID uuid.UUID)
		checkResponse func(t *testing.T, reassignResp handler.ReassignPullRequestResponse, prID uuid.UUID, oldReviewerID uuid.UUID, newReviewerID uuid.UUID)
	}{
		{
			name: "successful reassign pull request to different reviewer",
			setup: func(ctx context.Context, repos *TestRepos) (prID uuid.UUID, oldReviewerID uuid.UUID, newReviewerID uuid.UUID) {
				teamID := uuid.New()
				authorID := uuid.New()
				oldReviewerID = uuid.New()
				newReviewerID = uuid.New()
				prID = uuid.New()
				statusID := uuid.New()
				now := time.Now()

				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team For Reassign 1",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     authorID,
						Name:   "Author",
						TeamID: teamID,
					},
					{
						ID:       oldReviewerID,
						Name:     "Old Reviewer",
						TeamID:   teamID,
						IsActive: true,
					},
					{
						ID:       newReviewerID,
						Name:     "New Reviewer",
						TeamID:   teamID,
						IsActive: true,
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
					Name:      "PR For Reassignment",
					AuthorID:  authorID,
					StatusID:  statusID,
					CreatedAt: now,
				})
				assert.NoError(s.T(), err)

				_, err = repos.Reviewer.SavePRReviewer(ctx, pr_reviewers.PrReviewerIn{
					ID:         uuid.New(),
					PrID:       prID,
					ReviewerID: oldReviewerID,
				})
				assert.NoError(s.T(), err)

				return prID, oldReviewerID, newReviewerID
			},
			checkResponse: func(t *testing.T, reassignResp handler.ReassignPullRequestResponse, prID uuid.UUID, oldReviewerID uuid.UUID, newReviewerID uuid.UUID) {
				assert.NotNil(t, reassignResp)
				assert.Equal(t, prID, reassignResp.Pr.PullRequestId)

				assert.Contains(t, reassignResp.Pr.AssignedReviewers, newReviewerID)
				assert.NotContains(t, reassignResp.Pr.AssignedReviewers, oldReviewerID)
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

			prID, oldReviewerID, newReviewerID := tt.setup(ctx, repos)

			reassignResp, status, _, err := reassignPullRequest(token, prID, oldReviewerID)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, status)
			assert.NotNil(t, reassignResp)

			if tt.checkResponse != nil {
				tt.checkResponse(t, reassignResp, prID, oldReviewerID, newReviewerID)
			}
		})
	}
}
