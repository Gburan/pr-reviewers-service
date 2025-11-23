package add_team

import "github.com/google/uuid"

type In struct {
	TeamName string
	Members  []TeamMembers
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
