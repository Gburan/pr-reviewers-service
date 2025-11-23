package pull_requests

import (
	"time"

	"github.com/google/uuid"
)

type PullRequestIn struct {
	ID        uuid.UUID
	Name      string
	AuthorID  uuid.UUID
	StatusID  uuid.UUID
	CreatedAt time.Time
	MergedAt  time.Time
}

type PullRequestOut struct {
	ID        uuid.UUID
	Name      string
	AuthorID  uuid.UUID
	StatusID  uuid.UUID
	CreatedAt time.Time
	MergedAt  time.Time
}

type pullRequestDB struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	AuthorID  uuid.UUID `db:"author_id"`
	StatusID  uuid.UUID `db:"status_id"`
	CreatedAt time.Time `db:"created_at"`
	MergedAt  time.Time `db:"merged_at"`
}
