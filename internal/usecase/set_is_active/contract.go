package set_is_active

import "github.com/google/uuid"

type In struct {
	UserID   uuid.UUID
	IsActive bool
}

type Out struct {
	UserId   uuid.UUID
	Username string
	TeamName string
	IsActive bool
}
