package teams

import (
	"context"
	"testing"
	"time"

	nower2 "pr-reviewers-service/internal/infrastructure/nower"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *TeamsTest) TestSaveTeam() {
	tests := []struct {
		name        string
		input       TeamIn
		setup       func(ctx context.Context, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *TeamOut, expectedInput TeamIn)
	}{
		{
			name: "successful SaveTeam with given UUID",
			input: TeamIn{
				ID:        uuid.New(),
				Name:      "Test Team",
				CreatedAt: time.Now(),
			},
			setup: func(ctx context.Context, repo *Repository) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *TeamOut, expectedInput TeamIn) {
				assert.NotNil(t, result)
				assert.Equal(t, expectedInput.ID, result.ID)
				assert.Equal(t, expectedInput.Name, result.Name)
				assert.Equal(t, expectedInput.CreatedAt, result.CreatedAt)
			},
		},
		{
			name: "SaveTeam with zero UUID generates new one",
			input: TeamIn{
				Name: "Test Team Auto ID",
			},
			setup: func(ctx context.Context, repo *Repository) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *TeamOut, expectedInput TeamIn) {
				assert.NotNil(t, result)
				assert.NotEqual(t, uuid.Nil, result.ID)
				assert.Equal(t, expectedInput.Name, result.Name)
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name: "SaveTeam with zero CreatedAt uses current time",
			input: TeamIn{
				ID:   uuid.New(),
				Name: "Test Team Auto Time",
			},
			setup: func(ctx context.Context, repo *Repository) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *TeamOut, expectedInput TeamIn) {
				assert.NotNil(t, result)
				assert.False(t, result.CreatedAt.IsZero())
				assert.WithinDuration(t, time.Now(), result.CreatedAt, time.Second)
			},
		},
		{
			name: "SaveTeam with duplicate name returns error",
			input: TeamIn{
				ID:   uuid.New(),
				Name: "Duplicate Name Team",
			},
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SaveTeam(ctx, TeamIn{
					ID:   uuid.New(),
					Name: "Duplicate Name Team",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.Error,
			checkResult: func(t *testing.T, result *TeamOut, expectedInput TeamIn) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, repo)
			}

			result, err := repo.SaveTeam(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result, tt.input)
			}
		})
	}
}

func (s *TeamsTest) TestGetTeamByID() {
	teamID1 := uuid.New()
	teamID2 := uuid.New()

	tests := []struct {
		name        string
		input       uuid.UUID
		setup       func(ctx context.Context, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *TeamOut)
	}{
		{
			name:  "successful GetTeamByID returns team",
			input: teamID1,
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SaveTeam(ctx, TeamIn{
					ID:   teamID1,
					Name: "Test Team 1",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveTeam(ctx, TeamIn{
					ID:   teamID2,
					Name: "Test Team 2",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *TeamOut) {
				assert.NotNil(t, result)
				assert.Equal(t, teamID1, result.ID)
				assert.Equal(t, "Test Team 1", result.Name)
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name:  "GetTeamByID with non-existent ID returns not found error",
			input: uuid.New(),
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SaveTeam(ctx, TeamIn{
					ID:   teamID1,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
			checkResult: func(t *testing.T, result *TeamOut) {
				assert.Nil(t, result)
			},
		},
		{
			name:  "GetTeamByID with zero UUID returns not found error",
			input: uuid.Nil,
			setup: func(ctx context.Context, repo *Repository) {
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
			checkResult: func(t *testing.T, result *TeamOut) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, repo)
			}

			result, err := repo.GetTeamByID(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *TeamsTest) TestGetTeamByName() {
	teamID1 := uuid.New()
	teamID2 := uuid.New()

	tests := []struct {
		name        string
		input       string
		setup       func(ctx context.Context, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *TeamOut)
	}{
		{
			name:  "successful GetTeamByName returns team",
			input: "Test Team 1",
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SaveTeam(ctx, TeamIn{
					ID:   teamID1,
					Name: "Test Team 1",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SaveTeam(ctx, TeamIn{
					ID:   teamID2,
					Name: "Test Team 2",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *TeamOut) {
				assert.NotNil(t, result)
				assert.Equal(t, teamID1, result.ID)
				assert.Equal(t, "Test Team 1", result.Name)
				assert.False(t, result.CreatedAt.IsZero())
			},
		},
		{
			name:  "GetTeamByName is case sensitive",
			input: "test team 1",
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SaveTeam(ctx, TeamIn{
					ID:   teamID1,
					Name: "Test Team 1",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
			checkResult: func(t *testing.T, result *TeamOut) {
				assert.Nil(t, result)
			},
		},
		{
			name:  "GetTeamByName with non-existent name returns not found error",
			input: "Non-existent Team",
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SaveTeam(ctx, TeamIn{
					ID:   teamID1,
					Name: "Test Team",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
			checkResult: func(t *testing.T, result *TeamOut) {
				assert.Nil(t, result)
			},
		},
		{
			name:  "GetTeamByName with empty name returns not found error",
			input: "",
			setup: func(ctx context.Context, repo *Repository) {
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
			checkResult: func(t *testing.T, result *TeamOut) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repo := NewRepository(suite2.GlobalPool, nower2.Nower{})

			if tt.setup != nil {
				tt.setup(ctx, repo)
			}

			result, err := repo.GetTeamByName(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
