package repository

import "errors"

var (
	ErrBuildQuery          = errors.New("failed to build SQL query")
	ErrExecuteQuery        = errors.New("failed to execute query")
	ErrScanResult          = errors.New("failed to scan result")
	ErrUserNotFound        = errors.New("user not found")
	ErrTeamNotFound        = errors.New("team not found")
	ErrPullRequestNotFound = errors.New("pull request found")
	ErrPRStatusNotFound    = errors.New("pr status found")
	ErrPRReviewerNotFound  = errors.New("pr reviewer found")
)
