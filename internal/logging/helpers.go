package logging

import (
	"context"

	"github.com/google/uuid"
)

type logCtx struct {
	RequestID       uuid.UUID
	Status          int
	RequestStart    string
	RequestDuration string
	Method          string
	Path            string

	Limit            int
	Role             string
	Name             string
	TeamName         string
	TeamMembersCount int
	UserId           uuid.UUID
	AuthorId         uuid.UUID
	PullRequestId    uuid.UUID
}

func WithLogPullRequestID(ctx context.Context, prId uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.PullRequestId = prId
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{PullRequestId: prId})
}

func WithLogAuthorID(ctx context.Context, authorId uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.AuthorId = authorId
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{AuthorId: authorId})
}

func WithLogUserId(ctx context.Context, userId uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.UserId = userId
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{UserId: userId})
}

func WithLogTeamMembersCount(ctx context.Context, cnt int) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.TeamMembersCount = cnt
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{TeamMembersCount: cnt})
}

func WithLogTeamName(ctx context.Context, teamName string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.TeamName = teamName
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{TeamName: teamName})
}

func WithLogName(ctx context.Context, name string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Name = name
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Name: name})
}

func WithLogRole(ctx context.Context, role string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Role = role
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Role: role})
}

func WithLogRequestID(ctx context.Context, requestID uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.RequestID = requestID
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{RequestID: requestID})
}

func WithLogRequestPath(ctx context.Context, path string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Path = path
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Path: path})
}

func WithLogRequestMethod(ctx context.Context, method string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Method = method
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Method: method})
}

func WithLogRequestStatus(ctx context.Context, status int) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Status = status
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Status: status})
}

func WithLogRequestDuration(ctx context.Context, duration string) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.RequestDuration = duration
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{RequestDuration: duration})
}
