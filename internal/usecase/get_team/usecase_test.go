package get_team

import (
	"context"
	"errors"
	"testing"

	"pr-reviewers-service/internal/infrastructure/repository"
	teams2 "pr-reviewers-service/internal/infrastructure/repository/teams"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	usecase2 "pr-reviewers-service/internal/usecase"
	teams "pr-reviewers-service/internal/usecase/contract/repository/teams/mocks"
	users "pr-reviewers-service/internal/usecase/contract/repository/users/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	req := In{
		TeamName: "test-team",
	}

	teamOut := &teams2.TeamOut{
		ID:   teamID,
		Name: req.TeamName,
	}

	usersOut := []users2.UserOut{
		{
			ID:       userID1,
			Name:     "user1",
			IsActive: true,
			TeamID:   teamID,
		},
		{
			ID:       userID2,
			Name:     "user2",
			IsActive: false,
			TeamID:   teamID,
		},
	}

	tests := []struct {
		name      string
		req       In
		setupMock func(
			mockTeams *teams.MockRepositoryTeams,
			mockUsers *users.MockRepositoryUsers,
		)
		expected      *Out
		expectedError error
	}{
		{
			name: "successful get team with users",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), req.TeamName).
					Return(teamOut, nil)

				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&usersOut, nil)
			},
			expected: &Out{
				TeamName: req.TeamName,
				Members: []TeamMembers{
					{
						UserID:   userID1,
						Username: "user1",
						IsActive: true,
					},
					{
						UserID:   userID2,
						Username: "user2",
						IsActive: false,
					},
				},
			},
		},
		{
			name: "team not found",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), req.TeamName).
					Return(nil, repository.ErrTeamNotFound)
			},
			expectedError: usecase2.ErrTeamNotFound,
		},
		{
			name: "error getting team",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), req.TeamName).
					Return(nil, errors.New("database error"))
			},
			expectedError: usecase2.ErrGetTeam,
		},
		{
			name: "successful get team with no users",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), req.TeamName).
					Return(teamOut, nil)

				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(nil, repository.ErrUserNotFound)
			},
			expected: &Out{
				TeamName: req.TeamName,
				Members:  []TeamMembers{},
			},
		},
		{
			name: "error getting users",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), req.TeamName).
					Return(teamOut, nil)

				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(nil, errors.New("database error"))
			},
			expectedError: usecase2.ErrGetUsers,
		},
		{
			name: "successful get team with single user",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), req.TeamName).
					Return(teamOut, nil)

				singleUser := []users2.UserOut{
					{
						ID:       userID1,
						Name:     "single-user",
						IsActive: true,
						TeamID:   teamID,
					},
				}
				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&singleUser, nil)
			},
			expected: &Out{
				TeamName: req.TeamName,
				Members: []TeamMembers{
					{
						UserID:   userID1,
						Username: "single-user",
						IsActive: true,
					},
				},
			},
		},
		{
			name: "successful get team with inactive users only",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
			) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), req.TeamName).
					Return(teamOut, nil)

				inactiveUsers := []users2.UserOut{
					{
						ID:       userID1,
						Name:     "inactive-user1",
						IsActive: false,
						TeamID:   teamID,
					},
					{
						ID:       userID2,
						Name:     "inactive-user2",
						IsActive: false,
						TeamID:   teamID,
					},
				}
				mockUsers.EXPECT().
					GetUsersByTeamID(gomock.Any(), teamID).
					Return(&inactiveUsers, nil)
			},
			expected: &Out{
				TeamName: req.TeamName,
				Members: []TeamMembers{
					{
						UserID:   userID1,
						Username: "inactive-user1",
						IsActive: false,
					},
					{
						UserID:   userID2,
						Username: "inactive-user2",
						IsActive: false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoTeams := teams.NewMockRepositoryTeams(ctrl)
			mockRepoUsers := users.NewMockRepositoryUsers(ctrl)

			tt.setupMock(mockRepoTeams, mockRepoUsers)

			u := NewUsecase(mockRepoTeams, mockRepoUsers)
			result, err := u.Run(context.Background(), tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.expected != nil {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.TeamName, result.TeamName)
				assert.Equal(t, len(tt.expected.Members), len(result.Members))

				for i, expectedMember := range tt.expected.Members {
					assert.Equal(t, expectedMember.UserID, result.Members[i].UserID)
					assert.Equal(t, expectedMember.Username, result.Members[i].Username)
					assert.Equal(t, expectedMember.IsActive, result.Members[i].IsActive)
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
