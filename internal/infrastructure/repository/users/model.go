package users

import (
	"time"

	"github.com/google/uuid"
)

type UserIn struct {
	ID        uuid.UUID
	Name      string
	IsActive  bool
	TeamID    uuid.UUID
	CreatedAt time.Time
}

type UserOut struct {
	ID        uuid.UUID
	Name      string
	IsActive  bool
	TeamID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type userDB struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	IsActive  bool      `db:"is_active"`
	TeamID    uuid.UUID `db:"team_id"`
	CreatedAt time.Time `db:"created_at"`
}
