package add_team

import (
	"context"
	"errors"
	"testing"

	repository2 "pr-reviewers-service/internal/infrastructure/repository"
	teams2 "pr-reviewers-service/internal/infrastructure/repository/teams"
	users2 "pr-reviewers-service/internal/infrastructure/repository/users"
	usecase2 "pr-reviewers-service/internal/usecase"
	teams "pr-reviewers-service/internal/usecase/contract/repository/teams/mocks"
	users "pr-reviewers-service/internal/usecase/contract/repository/users/mocks"

	trmgr "github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/drivers/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	teamID := uuid.New()
	userID := uuid.New()

	reqData := In{
		TeamName: "team-1",
		Members: []TeamMembers{
			{
				UserID:   userID,
				Username: "user1",
				IsActive: true,
			},
		},
	}

	retTeam := &teams2.TeamOut{
		ID:   teamID,
		Name: reqData.TeamName,
	}
	retUser := users2.UserOut{
		ID:       userID,
		Name:     "user1",
		IsActive: true,
		TeamID:   teamID,
	}

	tests := []struct {
		name          string
		req           In
		setupMock     func(mockUsers *users.MockRepositoryUsers, mockTeams *teams.MockRepositoryTeams, trm trmgr.Manager)
		expected      *Out
		expectedError error
	}{
		{
			name: "successful add new team with new user",
			req:  reqData,
			setupMock: func(mockUsers *users.MockRepositoryUsers, mockTeams *teams.MockRepositoryTeams, trm trmgr.Manager) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), reqData.TeamName).
					Return(nil, repository2.ErrTeamNotFound)

				mockTeams.EXPECT().
					SaveTeam(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, teamIn teams2.TeamIn) (*teams2.TeamOut, error) {
						return retTeam, nil
					})

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), gomock.Any()).
					Return(nil, repository2.ErrUserNotFound)

				mockUsers.EXPECT().
					SaveUsersBatch(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, usersToCreate []users2.UserIn) (*[]users2.UserOut, error) {
						return &[]users2.UserOut{retUser}, nil
					})
			},
			expected: &Out{
				TeamName: reqData.TeamName,
				Members:  reqData.Members,
			},
		},
		{
			name: "successful update team user",
			req:  reqData,
			setupMock: func(mockUsers *users.MockRepositoryUsers, mockTeams *teams.MockRepositoryTeams, trm trmgr.Manager) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), reqData.TeamName).
					Return(retTeam, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{
						{
							ID:       userID,
							Name:     "diff_name",
							IsActive: true,
							TeamID:   teamID,
						},
					}, nil)

				mockUsers.EXPECT().
					UpdateUsersBatch(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, usersToCreate []users2.UserIn) (*[]users2.UserOut, error) {
						return &[]users2.UserOut{retUser}, nil
					})
			},
			expected: &Out{
				TeamName: reqData.TeamName,
				Members:  reqData.Members,
			},
		},
		{
			name: "duplicate users in request",
			req: In{
				TeamName: "team-1",
				Members: []TeamMembers{
					{UserID: userID, Username: "user1", IsActive: true},
					{UserID: userID, Username: "user1", IsActive: true},
				},
			},
			setupMock:     func(mockUsers *users.MockRepositoryUsers, mockTeams *teams.MockRepositoryTeams, trm trmgr.Manager) {},
			expectedError: usecase2.ErrDuplicateUsers,
		},
		{
			name: "team already exists",
			req:  reqData,
			setupMock: func(mockUsers *users.MockRepositoryUsers, mockTeams *teams.MockRepositoryTeams, trm trmgr.Manager) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), reqData.TeamName).
					Return(retTeam, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), gomock.Any()).
					Return(nil, repository2.ErrUserNotFound)

				mockUsers.EXPECT().
					SaveUsersBatch(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{retUser}, nil)
			},
			expected: &Out{
				TeamName: reqData.TeamName,
				Members:  reqData.Members,
			},
		},
		{
			name: "error on get team",
			req:  reqData,
			setupMock: func(mockUsers *users.MockRepositoryUsers, mockTeams *teams.MockRepositoryTeams, trm trmgr.Manager) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), reqData.TeamName).
					Return(nil, errors.New("some db error"))
			},
			expectedError: usecase2.ErrGetTeam,
		},
		{
			name: "error on save team",
			req:  reqData,
			setupMock: func(mockUsers *users.MockRepositoryUsers, mockTeams *teams.MockRepositoryTeams, trm trmgr.Manager) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), reqData.TeamName).
					Return(nil, repository2.ErrTeamNotFound)

				mockTeams.EXPECT().
					SaveTeam(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some db error"))
			},
			expectedError: usecase2.ErrSaveTeam,
		},
		{
			name: "no users were added or updated",
			req:  reqData,
			setupMock: func(mockUsers *users.MockRepositoryUsers, mockTeams *teams.MockRepositoryTeams, trm trmgr.Manager) {
				mockTeams.EXPECT().
					GetTeamByName(gomock.Any(), reqData.TeamName).
					Return(nil, repository2.ErrTeamNotFound)

				mockTeams.EXPECT().
					SaveTeam(gomock.Any(), gomock.Any()).
					Return(retTeam, nil)

				mockUsers.EXPECT().
					GetUsersByIDs(gomock.Any(), gomock.Any()).
					Return(&[]users2.UserOut{
						{
							ID:       userID,
							Name:     "user1",
							IsActive: true,
							TeamID:   teamID,
						},
					}, nil)
			},
			expectedError: usecase2.ErrNoUsersWereUpdatedAddedTeam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoUsers := users.NewMockRepositoryUsers(ctrl)
			mockRepoTeams := teams.NewMockRepositoryTeams(ctrl)

			mockTrm := mock.NewMockManager(ctrl)
			mockTrm.EXPECT().
				Do(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
					return f(ctx)
				}).AnyTimes()

			tt.setupMock(mockRepoUsers, mockRepoTeams, mockTrm)

			u := Newusecase(mockRepoUsers, mockRepoTeams, mockTrm)
			result, err := u.Run(context.Background(), tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorAs(t, err, &tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.expected != nil {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.TeamName, result.TeamName)
				assert.Equal(t, len(tt.expected.Members), len(result.Members))
				for i := range result.Members {
					assert.Equal(t, tt.expected.Members[i].UserID, result.Members[i].UserID)
					assert.Equal(t, tt.expected.Members[i].Username, result.Members[i].Username)
					assert.Equal(t, tt.expected.Members[i].IsActive, result.Members[i].IsActive)
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
