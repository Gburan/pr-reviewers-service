package set_is_active

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

	"github.com/avito-tech/go-transaction-manager/trm/v2/drivers/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetIsActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	teamID := uuid.New()

	req := In{
		UserID:   userID,
		IsActive: false,
	}
	existingUser := &users2.UserOut{
		ID:       userID,
		Name:     "test-user",
		IsActive: true,
		TeamID:   teamID,
	}
	team := &teams2.TeamOut{
		ID:   teamID,
		Name: "test-team",
	}
	updatedUser := &users2.UserOut{
		ID:       userID,
		Name:     "test-user",
		IsActive: false,
		TeamID:   teamID,
	}

	tests := []struct {
		name      string
		req       In
		setupMock func(
			mockTeams *teams.MockRepositoryTeams,
			mockUsers *users.MockRepositoryUsers,
			mockTrm *mock.MockManager,
		)
		expected      *Out
		expectedError error
	}{
		{
			name: "successful set user inactive",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockTrm *mock.MockManager,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(existingUser, nil)

				mockTeams.EXPECT().
					GetTeamByID(gomock.Any(), teamID).
					Return(team, nil)

				mockUsers.EXPECT().
					UpdateUser(gomock.Any(), users2.UserIn{
						ID:       userID,
						Name:     "test-user",
						IsActive: false,
						TeamID:   teamID,
					}).
					Return(updatedUser, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				UserId:   userID,
				Username: "test-user",
				TeamName: "test-team",
				IsActive: false,
			},
		},
		{
			name: "successful set user active",
			req: In{
				UserID:   userID,
				IsActive: true,
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockTrm *mock.MockManager,
			) {
				inactiveUser := &users2.UserOut{
					ID:       userID,
					Name:     "test-user",
					IsActive: false,
					TeamID:   teamID,
				}

				activatedUser := &users2.UserOut{
					ID:       userID,
					Name:     "test-user",
					IsActive: true,
					TeamID:   teamID,
				}

				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(inactiveUser, nil)

				mockTeams.EXPECT().
					GetTeamByID(gomock.Any(), teamID).
					Return(team, nil)

				mockUsers.EXPECT().
					UpdateUser(gomock.Any(), users2.UserIn{
						ID:       userID,
						Name:     "test-user",
						IsActive: true,
						TeamID:   teamID,
					}).
					Return(activatedUser, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expected: &Out{
				UserId:   userID,
				Username: "test-user",
				TeamName: "test-team",
				IsActive: true,
			},
		},
		{
			name: "user not found",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockTrm *mock.MockManager,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(nil, repository.ErrUserNotFound)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrUserNotFound,
		},
		{
			name: "error getting user",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockTrm *mock.MockManager,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrGetUser,
		},
		{
			name: "error getting team",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockTrm *mock.MockManager,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(existingUser, nil)

				mockTeams.EXPECT().
					GetTeamByID(gomock.Any(), teamID).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrGetTeam,
		},
		{
			name: "user already has required is_active value",
			req: In{
				UserID:   userID,
				IsActive: true,
			},
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockTrm *mock.MockManager,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(existingUser, nil)

				mockTeams.EXPECT().
					GetTeamByID(gomock.Any(), teamID).
					Return(team, nil)

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrUserDontNeedChange,
		},
		{
			name: "error updating user",
			req:  req,
			setupMock: func(
				mockTeams *teams.MockRepositoryTeams,
				mockUsers *users.MockRepositoryUsers,
				mockTrm *mock.MockManager,
			) {
				mockUsers.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(existingUser, nil)

				mockTeams.EXPECT().
					GetTeamByID(gomock.Any(), teamID).
					Return(team, nil)

				mockUsers.EXPECT().
					UpdateUser(gomock.Any(), users2.UserIn{
						ID:       userID,
						Name:     "test-user",
						IsActive: false,
						TeamID:   teamID,
					}).
					Return(nil, errors.New("database error"))

				mockTrm.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(ctx context.Context) error) error {
						return f(ctx)
					})
			},
			expectedError: usecase2.ErrUpdateUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepoTeams := teams.NewMockRepositoryTeams(ctrl)
			mockRepoUsers := users.NewMockRepositoryUsers(ctrl)
			mockTrm := mock.NewMockManager(ctrl)

			tt.setupMock(mockRepoTeams, mockRepoUsers, mockTrm)

			u := NewUsecase(mockRepoTeams, mockRepoUsers, mockTrm)
			result, err := u.Run(context.Background(), tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.expected != nil {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.UserId, result.UserId)
				assert.Equal(t, tt.expected.Username, result.Username)
				assert.Equal(t, tt.expected.TeamName, result.TeamName)
				assert.Equal(t, tt.expected.IsActive, result.IsActive)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
