package stats_pr_assignments

import "github.com/google/uuid"

type In struct{}

type Out struct {
	Reviewers []ReviewerStats
}

type ReviewerStats struct {
	ReviewerID      uuid.UUID
	AssignmentCount int
}
