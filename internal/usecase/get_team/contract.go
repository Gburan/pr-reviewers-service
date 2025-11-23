package get_team

import "github.com/google/uuid"

type In struct {
	TeamName string
}

type Out struct {
	TeamName string
	Members  []TeamMembers
}

type TeamMembers struct {
	IsActive bool
	UserID   uuid.UUID
	Username string
}
