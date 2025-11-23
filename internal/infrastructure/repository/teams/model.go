package teams

import (
	"time"

	"github.com/google/uuid"
)

type TeamIn struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type TeamOut struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type teamDB struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}
