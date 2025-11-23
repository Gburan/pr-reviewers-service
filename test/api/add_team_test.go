package api

import (
	"net/http"
	"testing"

	"pr-reviewers-service/internal/generated/api/v1/handler"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *ApiTest) TestAddTeam() {
	loginResp, status, _, err := dummyLogin()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status)
	token := loginResp.Token

	tests := []struct {
		name           string
		token          string
		teamName       string
		members        []handler.TeamMember
		expectedStatus int
		checkResponse  func(t *testing.T, team *handler.Team)
	}{
		{
			name:           "successful team creation - several active members",
			token:          token,
			teamName:       "team1",
			expectedStatus: http.StatusCreated,
			members: []handler.TeamMember{
				{
					UserId:   uuid.New(),
					Username: "developer1",
					IsActive: true,
				},
				{
					UserId:   uuid.New(),
					Username: "developer2",
					IsActive: true,
				},
				{
					UserId:   uuid.New(),
					Username: "developer3",
					IsActive: true,
				},
			},
			checkResponse: func(t *testing.T, team *handler.Team) {
				assert.NotNil(t, team)
				assert.Equal(t, "team1", team.TeamName)
				assert.Len(t, team.Members, 3)

				for _, member := range team.Members {
					assert.NotEqual(t, uuid.Nil, member.UserId)
					assert.NotEmpty(t, member.Username)
					assert.True(t, member.IsActive)
				}

				userIDs := make(map[uuid.UUID]bool)
				for _, member := range team.Members {
					assert.False(t, userIDs[member.UserId], "User IDs should be unique")
					userIDs[member.UserId] = true
				}
			},
		},
		{
			name:           "successful team creation - some users inactive",
			token:          token,
			teamName:       "team2",
			expectedStatus: http.StatusCreated,
			members: []handler.TeamMember{
				{
					UserId:   uuid.New(),
					Username: "developer1",
					IsActive: true,
				},
				{
					UserId:   uuid.New(),
					Username: "developer2",
					IsActive: false,
				},
			},
			checkResponse: func(t *testing.T, team *handler.Team) {
				assert.NotNil(t, team)
				assert.Equal(t, "team2", team.TeamName)
				assert.Len(t, team.Members, 2)

				assert.True(t, team.Members[0].IsActive)
				assert.Equal(t, "developer1", team.Members[0].Username)
				assert.NotEqual(t, uuid.Nil, team.Members[0].UserId)

				assert.False(t, team.Members[1].IsActive)
				assert.Equal(t, "developer2", team.Members[1].Username)
				assert.NotEqual(t, uuid.Nil, team.Members[1].UserId)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			team, status, _, err := addTeam(tt.token, tt.teamName, tt.members)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, status)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &team)
			}
		})
	}
}
