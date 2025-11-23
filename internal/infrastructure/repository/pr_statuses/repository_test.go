package pr_statuses

import (
	"context"
	"testing"

	"pr-reviewers-service/internal/infrastructure/repository"
	suite2 "pr-reviewers-service/test/suite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *PRStatusesTest) TestSavePRStatus() {
	duplicateID := uuid.New()

	tests := []struct {
		name        string
		input       PRStatusIn
		setup       func(ctx context.Context, repo *Repository, input PRStatusIn)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *PRStatusOut, expectedInput PRStatusIn)
	}{
		{
			name: "successful SavePRStatus with given UUID",
			input: PRStatusIn{
				ID:     uuid.New(),
				Status: "open",
			},
			setup: func(ctx context.Context, repo *Repository, input PRStatusIn) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PRStatusOut, expectedInput PRStatusIn) {
				assert.NotNil(t, result)
				assert.Equal(t, expectedInput.ID, result.ID)
				assert.Equal(t, expectedInput.Status, result.Status)
			},
		},
		{
			name: "SavePRStatus with zero UUID generates new one",
			input: PRStatusIn{
				Status: "closed",
			},
			setup: func(ctx context.Context, repo *Repository, input PRStatusIn) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PRStatusOut, expectedInput PRStatusIn) {
				assert.NotNil(t, result)
				assert.NotEqual(t, uuid.Nil, result.ID)
				assert.Equal(t, expectedInput.Status, result.Status)
			},
		},
		{
			name: "SavePRStatus with duplicate ID returns error",
			input: PRStatusIn{
				ID:     duplicateID,
				Status: "duplicate",
			},
			setup: func(ctx context.Context, repo *Repository, input PRStatusIn) {
				_, err := repo.SavePRStatus(ctx, PRStatusIn{
					ID:     duplicateID,
					Status: "original",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.Error,
			checkResult: func(t *testing.T, result *PRStatusOut, expectedInput PRStatusIn) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repo := NewRepository(suite2.GlobalPool)

			if tt.setup != nil {
				tt.setup(ctx, repo, tt.input)
			}

			result, err := repo.SavePRStatus(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result, tt.input)
			}
		})
	}
}

func (s *PRStatusesTest) TestGetPRStatusByID() {
	statusID1 := uuid.New()
	statusID2 := uuid.New()

	tests := []struct {
		name        string
		input       PRStatusIn
		setup       func(ctx context.Context, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *PRStatusOut)
	}{
		{
			name: "successful GetPRStatusByID returns status",
			input: PRStatusIn{
				ID: statusID1,
			},
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID1,
					Status: "open",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PRStatusOut) {
				assert.NotNil(t, result)
				assert.Equal(t, statusID1, result.ID)
				assert.Equal(t, "open", result.Status)
			},
		},
		{
			name: "GetPRStatusByID with non-existent ID returns not found error",
			input: PRStatusIn{
				ID: uuid.New(),
			},
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID2,
					Status: "closed",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err) &&
					assert.Contains(t, err.Error(), repository.ErrPRStatusNotFound.Error())
			},
			checkResult: func(t *testing.T, result *PRStatusOut) {
				assert.Nil(t, result)
			},
		},
		{
			name: "GetPRStatusByID with zero UUID returns not found error",
			input: PRStatusIn{
				ID: uuid.Nil,
			},
			setup: func(ctx context.Context, repo *Repository) {
			},
			checkErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err) &&
					assert.Contains(t, err.Error(), repository.ErrPRStatusNotFound.Error())
			},
			checkResult: func(t *testing.T, result *PRStatusOut) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repo := NewRepository(suite2.GlobalPool)

			if tt.setup != nil {
				tt.setup(ctx, repo)
			}

			result, err := repo.GetPRStatusByID(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *PRStatusesTest) TestGetPRStatusesByIDs() {
	statusID1 := uuid.New()
	statusID2 := uuid.New()
	statusID3 := uuid.New()
	statusID4 := uuid.New()

	tests := []struct {
		name        string
		input       []uuid.UUID
		setup       func(ctx context.Context, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *[]PRStatusOut)
	}{
		{
			name:  "successful GetPRStatusesByIDs returns multiple statuses",
			input: []uuid.UUID{statusID1, statusID2},
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID1,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID2,
					Status: "closed",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID3,
					Status: "merged",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PRStatusOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				statuses := *result
				statusMap := make(map[uuid.UUID]string)
				for _, status := range statuses {
					statusMap[status.ID] = status.Status
				}

				assert.Equal(t, "open", statusMap[statusID1])
				assert.Equal(t, "closed", statusMap[statusID2])
				assert.NotContains(t, statusMap, statusID3)
			},
		},
		{
			name:  "GetPRStatusesByIDs with empty list returns empty result",
			input: []uuid.UUID{},
			setup: func(ctx context.Context, repo *Repository) {
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PRStatusOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
		{
			name:  "GetPRStatusesByIDs with non-existent IDs returns empty result",
			input: []uuid.UUID{uuid.New(), uuid.New()},
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID4,
					Status: "draft",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PRStatusOut) {
				assert.NotNil(t, result)
				assert.Empty(t, *result)
			},
		},
		{
			name:  "GetPRStatusesByIDs with mixed existent and non-existent IDs",
			input: []uuid.UUID{statusID1, uuid.New(), statusID2},
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID1,
					Status: "open",
				})
				assert.NoError(s.T(), err)

				_, err = repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID2,
					Status: "closed",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *[]PRStatusOut) {
				assert.NotNil(t, result)
				assert.Len(t, *result, 2)

				statuses := *result
				statusMap := make(map[uuid.UUID]string)
				for _, status := range statuses {
					statusMap[status.ID] = status.Status
				}

				assert.Equal(t, "open", statusMap[statusID1])
				assert.Equal(t, "closed", statusMap[statusID2])
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repo := NewRepository(suite2.GlobalPool)

			if tt.setup != nil {
				tt.setup(ctx, repo)
			}

			result, err := repo.GetPRStatusesByIDs(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func (s *PRStatusesTest) TestUpdatePRStatusByID() {
	statusID1 := uuid.New()
	statusID2 := uuid.New()

	tests := []struct {
		name        string
		input       PRStatusIn
		setup       func(ctx context.Context, repo *Repository)
		checkErr    assert.ErrorAssertionFunc
		checkResult func(t *testing.T, result *PRStatusOut)
	}{
		{
			name: "successful UpdatePRStatusByID",
			input: PRStatusIn{
				ID:     statusID1,
				Status: "merged",
			},
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID1,
					Status: "open",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.NoError,
			checkResult: func(t *testing.T, result *PRStatusOut) {
				assert.NotNil(t, result)
				assert.Equal(t, statusID1, result.ID)
				assert.Equal(t, "merged", result.Status)
			},
		},
		{
			name: "UpdatePRStatusByID with non-existent ID returns error",
			input: PRStatusIn{
				ID:     uuid.New(),
				Status: "should_fail",
			},
			setup: func(ctx context.Context, repo *Repository) {
				_, err := repo.SavePRStatus(ctx, PRStatusIn{
					ID:     statusID2,
					Status: "closed",
				})
				assert.NoError(s.T(), err)
			},
			checkErr: assert.Error,
			checkResult: func(t *testing.T, result *PRStatusOut) {
				assert.Nil(t, result)
			},
		},
		{
			name: "UpdatePRStatusByID with zero UUID returns error",
			input: PRStatusIn{
				ID:     uuid.Nil,
				Status: "should_fail",
			},
			setup: func(ctx context.Context, repo *Repository) {
			},
			checkErr: assert.Error,
			checkResult: func(t *testing.T, result *PRStatusOut) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			ctx := context.Background()
			repo := NewRepository(suite2.GlobalPool)

			if tt.setup != nil {
				tt.setup(ctx, repo)
			}

			result, err := repo.UpdatePRStatusByID(ctx, tt.input)
			tt.checkErr(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			if err == nil && tt.checkResult != nil {
				updatedStatus, err := repo.GetPRStatusByID(ctx, PRStatusIn{ID: tt.input.ID})
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Status, updatedStatus.Status)
			}
		})
	}
}

func (s *PRStatusesTest) TestIntegration_SaveGetUpdateFlow() {
	s.T().Run("complete flow: save, get, update, get again", func(t *testing.T) {
		s.SetupTest()

		ctx := context.Background()
		repo := NewRepository(suite2.GlobalPool)

		statusID := uuid.New()
		saveInput := PRStatusIn{
			ID:     statusID,
			Status: "open",
		}

		savedStatus, err := repo.SavePRStatus(ctx, saveInput)
		assert.NoError(t, err)
		assert.Equal(t, statusID, savedStatus.ID)
		assert.Equal(t, "open", savedStatus.Status)

		retrievedStatus, err := repo.GetPRStatusByID(ctx, PRStatusIn{ID: statusID})
		assert.NoError(t, err)
		assert.Equal(t, savedStatus.ID, retrievedStatus.ID)
		assert.Equal(t, savedStatus.Status, retrievedStatus.Status)

		updateInput := PRStatusIn{
			ID:     statusID,
			Status: "closed",
		}

		updatedStatus, err := repo.UpdatePRStatusByID(ctx, updateInput)
		assert.NoError(t, err)
		assert.Equal(t, statusID, updatedStatus.ID)
		assert.Equal(t, "closed", updatedStatus.Status)

		finalStatus, err := repo.GetPRStatusByID(ctx, PRStatusIn{ID: statusID})
		assert.NoError(t, err)
		assert.Equal(t, statusID, finalStatus.ID)
		assert.Equal(t, "closed", finalStatus.Status)

		statusID2 := uuid.New()
		_, err = repo.SavePRStatus(ctx, PRStatusIn{
			ID:     statusID2,
			Status: "merged",
		})
		assert.NoError(t, err)

		statuses, err := repo.GetPRStatusesByIDs(ctx, []uuid.UUID{statusID, statusID2})
		assert.NoError(t, err)
		assert.Len(t, *statuses, 2)

		statusMap := make(map[uuid.UUID]string)
		for _, status := range *statuses {
			statusMap[status.ID] = status.Status
		}

		assert.Equal(t, "closed", statusMap[statusID])
		assert.Equal(t, "merged", statusMap[statusID2])
	})
}
