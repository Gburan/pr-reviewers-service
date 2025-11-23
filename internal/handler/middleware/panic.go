package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"pr-reviewers-service/internal/generated/api/v1/handler"
)

func PanicMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if err := recover(); err != nil {
			slog.Error(
				"Caught panic",
				"method", r.Method,
				"url", r.URL.Path,
				"time", time.Since(start),
				"stack trace", string(debug.Stack()),
			)

			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(handler.ErrorResponse{
				Error: struct {
					Code    handler.ErrorResponseErrorCode `json:"code"`
					Message string                         `json:"message"`
				}{
					Code:    handler.UNKNOWN,
					Message: "Internal server error",
				},
			})
		}
		next.ServeHTTP(w, r)
	})
}
