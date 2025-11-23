package users

import (
	"context"

	"pr-reviewers-service/internal/infrastructure/repository/users"

	"github.com/google/uuid"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=users RepositoryUsers
type RepositoryUsers interface {
	GetUserByID(ctx context.Context, userId uuid.UUID) (*users.UserOut, error)
	GetUsersByIDs(ctx context.Context, userIds []uuid.UUID) (*[]users.UserOut, error)
	GetUsersByTeamID(ctx context.Context, teamID uuid.UUID) (*[]users.UserOut, error)
	GetActiveUsersByTeamID(ctx context.Context, teamID uuid.UUID) (*[]users.UserOut, error)
	UpdateUser(ctx context.Context, user users.UserIn) (*users.UserOut, error)
	SaveUsersBatch(ctx context.Context, urs []users.UserIn) (*[]users.UserOut, error)
	UpdateUsersBatch(ctx context.Context, urs []users.UserIn) (*[]users.UserOut, error)
}
