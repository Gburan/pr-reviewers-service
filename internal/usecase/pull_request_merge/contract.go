package pull_request_merge

import (
	"time"

	"github.com/google/uuid"
)

type In struct {
	PullRequestID uuid.UUID
}

type Out struct {
	PullRequestID     uuid.UUID
	PullRequestName   string
	AuthorID          uuid.UUID
	Status            string
	AssignedReviewers []uuid.UUID
	CreatedAt         time.Time
	MergedAt          time.Time
}
