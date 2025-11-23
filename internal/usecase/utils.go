package usecase

import "errors"

const (
	MergedStatusValue = "MERGED"
	OpenStatusValue   = "OPEN"
)

var (
	ErrGetTeam                     = errors.New("failed to get team")
	ErrSaveTeam                    = errors.New("failed to save team")
	ErrGetUsers                    = errors.New("failed to get users")
	ErrGetUser                     = errors.New("failed to get user")
	ErrGetPRStatus                 = errors.New("failed to get pr status")
	ErrGetPRReviewers              = errors.New("failed to get assigned reviewers")
	ErrPRsReviewersNotFound        = errors.New("prs reviewers found")
	ErrSaveUsersBatch              = errors.New("failed to save users batch")
	ErrSetPRStatus                 = errors.New("failed to save pr status")
	ErrUpdatePrMergeTime           = errors.New("failed to update pr merge time")
	ErrUpdatePrStatus              = errors.New("failed to update pr status")
	ErrGetPullRequest              = errors.New("failed to get pull request")
	ErrSavePullRequest             = errors.New("failed to save pull request")
	ErrRemoveReviewer              = errors.New("failed to remove pr reviewera")
	ErrAssignReviewer              = errors.New("failed to assign reviewer")
	ErrUpdateUsersBatch            = errors.New("failed to update users batch")
	ErrUpdateUser                  = errors.New("failed to update user")
	ErrPullRequestNotFound         = errors.New("not found such pull request")
	ErrNoUsersWereUpdatedAddedTeam = errors.New("no one was added to team")
	ErrAuthorPrNotFound            = errors.New("not found such user try to create pr from")
	ErrUserDontNeedChange          = errors.New("no need to change user")
	ErrNoAvailableReviewers        = errors.New("no available users")
	ErrNoActiveReviewers           = errors.New("no active reviewers at this pr")
	ErrPullRequestExists           = errors.New("such pr already exist")
	ErrPullRequestAlreadyMerged    = errors.New("such pr already merged")
	ErrDuplicateUsers              = errors.New("duplicate users ids got")
	ErrReviewerNotFound            = errors.New("not found such reviewer for this pr")
	ErrTeamNotFound                = errors.New("team not found")
	ErrUserNotFound                = errors.New("user not found")
	ErrUsersByIDsNotFound          = errors.New("not found user by ids in request")
	ErrUserNotBelongsToTeam        = errors.New("user not belongs to team")
	ErrNoPRsToAffect               = errors.New("no prs to change")
	ErrNoUsersAssignedToPRs        = errors.New("no users that assign to prs")
)
