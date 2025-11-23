package pr_statuses

import "github.com/google/uuid"

type PRStatusIn struct {
	ID     uuid.UUID
	Status string
}

type PRStatusOut struct {
	ID     uuid.UUID
	Status string
}

type prStatusDB struct {
	ID     uuid.UUID `db:"id"`
	Status string    `db:"status"`
}
