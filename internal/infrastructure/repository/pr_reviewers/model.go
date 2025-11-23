package pr_reviewers

import "github.com/google/uuid"

type PrReviewerOut struct {
	ID         uuid.UUID
	PRID       uuid.UUID
	ReviewerID uuid.UUID
}

type PrReviewerIn struct {
	ID         uuid.UUID
	PrID       uuid.UUID
	ReviewerID uuid.UUID
}

type prReviewerDB struct {
	ID         uuid.UUID `db:"id"`
	PRID       uuid.UUID `db:"pr_id"`
	ReviewerID uuid.UUID `db:"reviewer_id"`
}
