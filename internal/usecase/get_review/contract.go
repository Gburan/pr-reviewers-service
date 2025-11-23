package get_review

import (
	"github.com/google/uuid"
)

type In struct {
	UserID uuid.UUID
}

type Out struct {
	UserID       uuid.UUID
	PullRequests []PullRequestShort
}

type PullRequestShort struct {
	PullRequestID   uuid.UUID
	PullRequestName string
	AuthorID        uuid.UUID
	Status          string
}
