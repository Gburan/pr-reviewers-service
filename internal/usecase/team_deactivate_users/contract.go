package team_deactivate_users

import "github.com/google/uuid"

type In struct {
	TeamName string
	UserIDs  []uuid.UUID
}

type TeamMember struct {
	UserID   uuid.UUID
	Username string
	IsActive bool
}

type Team struct {
	TeamName string
	Members  []TeamMember
}

type PullRequestShort struct {
	PullRequestID   uuid.UUID
	PullRequestName string
	AuthorID        uuid.UUID
	Status          string
}

type Out struct {
	Team                 Team
	AffectedPullRequests []PullRequestShort
}
