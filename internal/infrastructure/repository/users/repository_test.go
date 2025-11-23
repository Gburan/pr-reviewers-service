package users

import (
	"context"
	"testing"

	nower2 "pr-reviewers-service/internal/infrastructure/nower"
	"pr-reviewers-service/internal/infrastructure/repository/teams"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *UsersTest) TestGetUserByID() {
	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	tests := []struct {
		name        string
		input       uuid.UUID
		setup       func(ctx context.Context, teamRepo *teams.Repository, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *UserOut)
	}{
		{
			name:  "successful GetUserByID returns user",
			input: userID1,
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "User 1",
						IsActive: true,
						TeamID:   teamID,
					},
					{
						ID:       userID2,
						Name:     "User 2",
						IsActive: false,
						TeamID:   teamID,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *UserOut) {
				assert.NotNil(t, result)
				assert.Equal(t, userID1, result.ID)
				assert.Equal(t, "User 1", result.Name)
				assert.True(t, result.IsActive)
				assert.Equal(t, teamID, result.TeamID)
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name:  "GetUserByID with non-existent ID returns not found error",
			input: uuid.New(),
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "User 1",
						IsActive: true,
						TeamID:   teamID,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
			checkResult: func(t *testing.T, result *UserOut) {
				assert.Nil(t, result)
			},
		},
		{
			name:  "GetUserByID with zero UUID returns not found error",
			input: uuid.Nil,
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
			checkResult: func(t *testing.T, result *UserOut) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			teamRepo := teams.NewRepository(suite2.GlobalPool, nower2.Nower{})
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, teamRepo, repo)
			}

			result, err := repo.GetUserByID(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *UsersTest) TestGetUsersByIDs() {
	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	tests := []struct {
		name        string
		input       []uuid.UUID
		setup       func(ctx context.Context, teamRepo *teams.Repository, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]UserOut)
	}{
		{
			name:  "successful GetUsersByIDs returns multiple users",
			input: []uuid.UUID{userID1, userID2},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "User 1",
						IsActive: true,
						TeamID:   teamID,
					},
					{
						ID:       userID2,
						Name:     "User 2",
						IsActive: false,
						TeamID:   teamID,
					},
					{
						ID:       userID3,
						Name:     "User 3",
						IsActive: true,
						TeamID:   teamID,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				users := *result
				userMap := make(map[uuid.UUID]UserOut)
				for _, user := range users {
					userMap[user.ID] = user
				}

				assert.Equal(t, "User 1", userMap[userID1].Name)
				assert.True(t, userMap[userID1].IsActive)
				assert.Equal(t, "User 2", userMap[userID2].Name)
				assert.False(t, userMap[userID2].IsActive)
				assert.NotContains(t, userMap, userID3)
			},
		},
		{
			name:  "GetUsersByIDs with empty list returns empty result",
			input: []uuid.UUID{},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
		{
			name:  "GetUsersByIDs with non-existent IDs returns empty result",
			input: []uuid.UUID{uuid.New(), uuid.New()},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "User 1",
						IsActive: true,
						TeamID:   teamID,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			teamRepo := teams.NewRepository(suite2.GlobalPool, nower2.Nower{})
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, teamRepo, repo)
			}

			result, err := repo.GetUsersByIDs(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *UsersTest) TestGetUsersByTeamID() {
	teamID1 := uuid.New()
	teamID2 := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	tests := []struct {
		name        string
		input       uuid.UUID
		setup       func(ctx context.Context, teamRepo *teams.Repository, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]UserOut)
	}{
		{
			name:  "successful GetUsersByTeamID returns team users",
			input: teamID1,
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID1,
					Name: "Team 1",
				})
				assert.NoError(s.T(), err)

				_, err = teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID2,
					Name: "Team 2",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "User 1 Team 1",
						IsActive: true,
						TeamID:   teamID1,
					},
					{
						ID:       userID2,
						Name:     "User 2 Team 1",
						IsActive: false,
						TeamID:   teamID1,
					},
					{
						ID:       userID3,
						Name:     "User 3 Team 2",
						IsActive: true,
						TeamID:   teamID2,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				users := *result
				userNames := make(map[string]bool)
				for _, user := range users {
					userNames[user.Name] = true
					assert.Equal(t, teamID1, user.TeamID)
				}

				assert.True(t, userNames["User 1 Team 1"])
				assert.True(t, userNames["User 2 Team 1"])
				assert.False(t, userNames["User 3 Team 2"])
			},
		},
		{
			name:  "GetUsersByTeamID with non-existent team returns empty result",
			input: uuid.New(),
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID1,
					Name: "Team 1",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "User 1",
						IsActive: true,
						TeamID:   teamID1,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			teamRepo := teams.NewRepository(suite2.GlobalPool, nower2.Nower{})
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, teamRepo, repo)
			}

			result, err := repo.GetUsersByTeamID(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *UsersTest) TestGetActiveUsersByTeamID() {
	teamID1 := uuid.New()
	teamID2 := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	tests := []struct {
		name        string
		input       uuid.UUID
		setup       func(ctx context.Context, teamRepo *teams.Repository, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]UserOut)
	}{
		{
			name:  "successful GetActiveUsersByTeamID returns only active users",
			input: teamID1,
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID1,
					Name: "Team 1",
				})
				assert.NoError(s.T(), err)

				_, err = teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID2,
					Name: "Team 2",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "Active User 1",
						IsActive: true,
						TeamID:   teamID1,
					},
					{
						ID:       userID2,
						Name:     "Inactive User",
						IsActive: false,
						TeamID:   teamID1,
					},
					{
						ID:       userID3,
						Name:     "Active User 2",
						IsActive: true,
						TeamID:   teamID1,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				users := *result
				for _, user := range users {
					assert.True(t, user.IsActive)
					assert.Equal(t, teamID1, user.TeamID)
				}
			},
		},
		{
			name:  "GetActiveUsersByTeamID with no active users returns empty result",
			input: teamID2,
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID2,
					Name: "Team 2",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "Inactive User 1",
						IsActive: false,
						TeamID:   teamID2,
					},
					{
						ID:       userID2,
						Name:     "Inactive User 2",
						IsActive: false,
						TeamID:   teamID2,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			teamRepo := teams.NewRepository(suite2.GlobalPool, nower2.Nower{})
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, teamRepo, repo)
			}

			result, err := repo.GetActiveUsersByTeamID(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *UsersTest) TestUpdateUser() {
	teamID1 := uuid.New()
	teamID2 := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		input       UserIn
		setup       func(ctx context.Context, teamRepo *teams.Repository, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *UserOut)
	}{
		{
			name: "successful UpdateUser",
			input: UserIn{
				ID:       userID,
				Name:     "Updated User",
				IsActive: false,
				TeamID:   teamID2,
			},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID1,
					Name: "Team 1",
				})
				assert.NoError(s.T(), err)

				_, err = teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID2,
					Name: "Team 2",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID,
						Name:     "Original User",
						IsActive: true,
						TeamID:   teamID1,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *UserOut) {
				assert.NotNil(t, result)
				assert.Equal(t, userID, result.ID)
				assert.Equal(t, "Updated User", result.Name)
				assert.False(t, result.IsActive)
				assert.Equal(t, teamID2, result.TeamID)
			},
		},
		{
			name: "UpdateUser with non-existent ID returns error",
			input: UserIn{
				ID:       uuid.New(),
				Name:     "Non-existent User",
				IsActive: true,
				TeamID:   teamID1,
			},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID1,
					Name: "Team 1",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.Error,
			checkResult: func(t *testing.T, result *UserOut) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			teamRepo := teams.NewRepository(suite2.GlobalPool, nower2.Nower{})
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, teamRepo, repo)
			}

			result, err := repo.UpdateUser(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *UsersTest) TestSaveUsersBatch() {
	teamID := uuid.New()

	tests := []struct {
		name        string
		input       []UserIn
		setup       func(ctx context.Context, teamRepo *teams.Repository, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]UserOut)
	}{
		{
			name: "successful SaveUsersBatch with multiple users",
			input: []UserIn{
				{
					ID:       uuid.New(),
					Name:     "User 1",
					IsActive: true,
					TeamID:   teamID,
				},
				{
					Name:     "User 2 Auto ID",
					IsActive: false,
					TeamID:   teamID,
				},
				{
					ID:       uuid.New(),
					Name:     "User 3",
					IsActive: true,
					TeamID:   teamID,
				},
			},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 3)

				users := *result
				userMap := make(map[string]UserOut)
				for _, user := range users {
					userMap[user.Name] = user
					assert.NotEqual(t, uuid.Nil, user.ID)
					assert.Equal(t, teamID, user.TeamID)
					assert.False(t, user.CreatedAt.IsZero())
				}

				assert.Equal(t, "User 1", userMap["User 1"].Name)
				assert.True(t, userMap["User 1"].IsActive)
				assert.Equal(t, "User 2 Auto ID", userMap["User 2 Auto ID"].Name)
				assert.False(t, userMap["User 2 Auto ID"].IsActive)
				assert.Equal(t, "User 3", userMap["User 3"].Name)
				assert.True(t, userMap["User 3"].IsActive)

				createdAt := users[0].CreatedAt
				for _, user := range users {
					assert.Equal(t, createdAt, user.CreatedAt)
				}
			},
		},
		{
			name:  "SaveUsersBatch with empty list returns empty result",
			input: []UserIn{},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			teamRepo := teams.NewRepository(suite2.GlobalPool, nower2.Nower{})
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, teamRepo, repo)
			}

			result, err := repo.SaveUsersBatch(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *UsersTest) TestUpdateUsersBatch() {
	teamID1 := uuid.New()
	teamID2 := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	tests := []struct {
		name        string
		input       []UserIn
		setup       func(ctx context.Context, teamRepo *teams.Repository, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]UserOut)
	}{
		{
			name: "successful UpdateUsersBatch with multiple users",
			input: []UserIn{
				{
					ID:       userID1,
					Name:     "Updated User 1",
					IsActive: false,
					TeamID:   teamID2,
				},
				{
					ID:       userID2,
					Name:     "Updated User 2",
					IsActive: true,
					TeamID:   teamID1,
				},
			},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
				_, err := teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID1,
					Name: "Team 1",
				})
				assert.NoError(s.T(), err)

				_, err = teamRepo.SaveTeam(ctx, teams.TeamIn{
					ID:   teamID2,
					Name: "Team 2",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveUsersBatch(ctx, []UserIn{
					{
						ID:       userID1,
						Name:     "Original User 1",
						IsActive: true,
						TeamID:   teamID1,
					},
					{
						ID:       userID2,
						Name:     "Original User 2",
						IsActive: false,
						TeamID:   teamID1,
					},
					{
						ID:       userID3,
						Name:     "User 3",
						IsActive: true,
						TeamID:   teamID1,
					},
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				users := *result
				userMap := make(map[uuid.UUID]UserOut)
				for _, user := range users {
					userMap[user.ID] = user
				}

				assert.Equal(t, "Updated User 1", userMap[userID1].Name)
				assert.False(t, userMap[userID1].IsActive)
				assert.Equal(t, teamID2, userMap[userID1].TeamID)

				assert.Equal(t, "Updated User 2", userMap[userID2].Name)
				assert.True(t, userMap[userID2].IsActive)
				assert.Equal(t, teamID1, userMap[userID2].TeamID)
			},
		},
		{
			name:  "UpdateUsersBatch with empty list returns empty result",
			input: []UserIn{},
			setup: func(ctx context.Context, teamRepo *teams.Repository, repo *Repository) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]UserOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			teamRepo := teams.NewRepository(suite2.GlobalPool, nower2.Nower{})
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, teamRepo, repo)
			}

			result, err := repo.UpdateUsersBatch(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			if err == nil && len(tt.input) > 0 {
				userIDs := make([]uuid.UUID, len(tt.input))
				for i, user := range tt.input {
					userIDs[i] = user.ID
				}

				updatedUsers, err := repo.GetUsersByIDs(ctx, userIDs)
				assert.NoError(t, err)
				assert.Len(t, *updatedUsers, len(tt.input))

				for _, expectedUser := range tt.input {
					for _, actualUser := range *updatedUsers {
						if actualUser.ID == expectedUser.ID {
							assert.Equal(t, expectedUser.Name, actualUser.Name)
							assert.Equal(t, expectedUser.IsActive, actualUser.IsActive)
							assert.Equal(t, expectedUser.TeamID, actualUser.TeamID)
						}
					}
				}
			}
		})
	}
}
