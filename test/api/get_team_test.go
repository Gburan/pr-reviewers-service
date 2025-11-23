package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	nower2 "pr-reviewers-service/internal/infrastructure/nower"
	"pr-reviewers-service/internal/infrastructure/repository/teams"
	"pr-reviewers-service/internal/infrastructure/repository/users"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *ApiTest) TestGetTeam() {
	loginResp, status, _, err := dummyLogin()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)
	token := loginResp.Token

	type TestRepos struct {
		Team *teams.Repository
		User *users.Repository
	}

	tests := []struct {
		name          string
		setup         func(ctx context.Context, repos *TestRepos) (teamName string, expectedTeam handler.Team)
		checkResponse func(t *testing.T, teamResp handler.Team, expectedTeam handler.Team)
	}{
		{
			name: "successful get team with multiple members",
			setup: func(ctx context.Context, repos *TestRepos) (teamName string, expectedTeam handler.Team) {
				teamID := uuid.New()
				teamName = "Development-Team"
				user1ID := uuid.New()
				user2ID := uuid.New()
				user3ID := uuid.New()

				teamOut, err := repos.Team.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: teamName,
				})
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), teamID, teamOut.ID)

				usersOut, err := repos.User.SaveUsersBatch(ctx, []users.UserIn{
					{
						ID:       user1ID,
						Name:     "Developer 1",
						TeamID:   teamID,
						IsActive: true,
					},
					{
						ID:     user2ID,
						Name:   "Developer 2",
						TeamID: teamID,
					},
					{
						ID:       user3ID,
						Name:     "Team Lead",
						TeamID:   teamID,
						IsActive: true,
					},
				})
				assert.NoError(s.T(), err)
				assert.Len(s.T(), *usersOut, 3)

				expectedTeam = handler.Team{
					TeamName: teamName,
					Members: []handler.TeamMember{
						{
							UserId:   user1ID,
							Username: "Developer 1",
							IsActive: true,
						},
						{
							UserId:   user2ID,
							Username: "Developer 2",
							IsActive: false,
						},
						{
							UserId:   user3ID,
							Username: "Team Lead",
							IsActive: true,
						},
					},
				}
				return teamName, expectedTeam
			},
			checkResponse: func(t *testing.T, teamResp handler.Team, expectedTeam handler.Team) {
				assert.NotNil(t, teamResp)
				assert.Equal(t, expectedTeam.TeamName, teamResp.TeamName)
				assert.Len(t, teamResp.Members, len(expectedTeam.Members))

				expectedMembers := make(map[uuid.UUID]handler.TeamMember)
				for _, member := range expectedTeam.Members {
					expectedMembers[member.UserId] = member
				}

				for _, member := range teamResp.Members {
					expectedMember, exists := expectedMembers[member.UserId]
					assert.True(t, exists, "Unexpected member %s in response", member.UserId)
					assert.Equal(t, expectedMember.Username, member.Username)
					assert.Equal(t, expectedMember.IsActive, member.IsActive)
				}
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			repos := &TestRepos{
				Team: teams.NewRepository(suite2.GlobalPool, nower2.Nower{}),
				User: users.NewRepository(suite2.GlobalPool, nower2.Nower{}),
			}

			teamName, expectedTeam := tt.setup(ctx, repos)

			teamResp, status, kal, err := getTeam(token, teamName)
			fmt.Println(kal)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, status)

			if tt.checkResponse != nil {
				tt.checkResponse(t, teamResp, expectedTeam)
			}
		})
	}
}
