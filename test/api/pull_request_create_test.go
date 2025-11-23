package api

import (
	"context"
	"net/http"
	"testing"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	nower2 "pr-reviewers-service/internal/infrastructure/nower"
	"pr-reviewers-service/internal/infrastructure/repository/pr_statuses"
	"pr-reviewers-service/internal/infrastructure/repository/teams"
	"pr-reviewers-service/internal/infrastructure/repository/users"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *ApiTest) TestPullRequestCreate() {
	loginResp, status, _, err := dummyLogin()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)
	token := loginResp.Token

	type TestRepos struct {
		Team   *teams.Repository
		User   *users.Repository
		Status *pr_statuses.Repository
	}

	tests := []struct {
		name  string
		setup func(ctx context.Context, repos *TestRepos) (userID uuid.UUID, prID uuid.UUID,
			prName string, expectedStatus string)
		checkResponse func(t *testing.T, prResp *handler.CreatePullRequestResponse,
			expectedUserID uuid.UUID, expectedPrID uuid.UUID, expectedPrName string, expectedStatus string)
	}{
		{
			name: "successful PR creation with valid data",
			setup: func(ctx context.Context, repos *TestRepos) (userID uuid.UUID, prID uuid.UUID,
				prName string, expectedStatus string) {
				teamID := uuid.New()
				userID = uuid.New()
				prID = uuid.New()
				prName = "Test PR 1"
				statusID := uuid.New()
				expectedStatus = "OPEN"

				_, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Development Team",
				})
				assert.NoError(s.T(), err)

				_, err = repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:     userID,
						Name:   "Test Author",
						TeamID: teamID,
					},
				})
				assert.NoError(s.T(), err)

				_, err = repos.Status.SavePRStatus(ctx, pr_statuses.PRStatusIn{
					ID:     statusID,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				return userID, prID, prName, expectedStatus
			},
			checkResponse: func(t *testing.T, prResp *handler.CreatePullRequestResponse, expectedUserID uuid.UUID,
				expectedPrID uuid.UUID, expectedPrName string, expectedStatus string) {
				assert.NotNil(t, prResp)
				assert.Equal(t, expectedPrName, prResp.Pr.PullRequestName)
				assert.Equal(t, expectedPrID, prResp.Pr.PullRequestId)
				assert.Equal(t, expectedUserID, prResp.Pr.AuthorId)
				assert.Equal(t, handler.PullRequestStatus(expectedStatus), prResp.Pr.Status)
				assert.False(t, prResp.Pr.CreatedAt.IsZero())
				assert.Nil(t, prResp.Pr.MergedAt)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			repos := &TestRepos{
				Team:   teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User:   users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				Status: pr_statuses.NewRepository(suite2.GlobalPool),
			}

			userID, prID, prName, expectedStatus := tt.setup(ctx, repos)

			prResp, status, _, err := createPullRequest(token, userID, prID, prName)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, status)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &prResp, userID, prID, prName, expectedStatus)
			}
		})
	}
}
