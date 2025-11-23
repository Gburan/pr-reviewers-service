package middleware

import (
	"errors"
	"net/http"
	"strings"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/jwt"
)

var (
	ErrNoToken          = errors.New("no token provided")
	ErrNoAcceptableRole = errors.New("no acceptable role")
)

type UserRole string

const (
	User  UserRole = "USER"
	Admin UserRole = "ADMIN"

	authorisationPrefix = "Bearer "
)

func hasRequiredRole(userRole UserRole, requiredRoles []UserRole) bool {
	role := UserRole(strings.ToUpper(string(userRole)))
	for _, requiredRole := range requiredRoles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoToken
	}

	return strings.TrimPrefix(authHeader, authorisationPrefix), nil
}

func AuthMiddleware(secret string, requiredRoles []UserRole, next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := extractToken(r)
		if err != nil {
			handler.RespondWithError(w, r.Context(), http.StatusUnauthorized, handler2.UNKNOWN, "authorization required", err)
			return
		}

		role, err := jwt.ParseToken(tokenString, secret)
		if err != nil {
			handler.RespondWithError(w, r.Context(), http.StatusUnauthorized, handler2.UNKNOWN, "invalid token", err)
			return
		}

		userRole := UserRole(role)
		if !hasRequiredRole(userRole, requiredRoles) {
			handler.RespondWithError(w, r.Context(), http.StatusForbidden, handler2.UNKNOWN, "insufficient permissions", ErrNoAcceptableRole)
			return
		}

		next.ServeHTTP(w, r)
	})
}
