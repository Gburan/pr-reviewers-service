package pull_request_reassign

import (
	"time"

	"github.com/google/uuid"
)

type In struct {
	PullRequestID uuid.UUID
	OldUserId     uuid.UUID
}

type Out struct {
	PullRequestID     uuid.UUID
	PullRequestName   string
	AuthorID          uuid.UUID
	Status            string
	AssignedReviewers []uuid.UUID
	CreatedAt         time.Time
	MergedAt          time.Time
	ReplacedBy        uuid.UUID
}
