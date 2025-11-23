package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/logging"
)

func RespondWithError(w http.ResponseWriter, ctx context.Context, status int, statusErrCode handler.ErrorResponseErrorCode, errorMsg string, err error) {
	if status == http.StatusInternalServerError {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), fmt.Sprintf("Error: %s", err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := handler.ErrorResponse{
		Error: struct {
			Code    handler.ErrorResponseErrorCode `json:"code"`
			Message string                         `json:"message"`
		}{
			Code: statusErrCode,
			Message: func() string {
				if err != nil {
					return fmt.Sprintf("%s: %s", errorMsg, err.Error())
				}
				return errorMsg
			}(),
		},
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(ctx, "Failed to encode error response", "error", err)
	}
}
