package dummy_login

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler"
	"pr-reviewers-service/internal/handler/middleware"
	"pr-reviewers-service/internal/jwt"
	"pr-reviewers-service/internal/logging"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	expIn = time.Duration(math.MaxInt64)
)

var (
	dummyId = uuid.New()
)

type createHandler struct {
	secret    string
	validator *validator.Validate
}

func New(secret string, validator *validator.Validate) *createHandler {
	return &createHandler{
		secret:    secret,
		validator: validator,
	}
}

// @Summary Dummy login
// @Description Get JWT token for testing purposes
// @ID DummyLogin
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} handler2.DummyLoginOut "Successfully logged in"
// @Failure 500 {object} handler2.ErrorResponse "Internal server error"
// @Router /dummyLogin [post]
func (h *createHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	token, err := jwt.GenerateToken(h.secret, string(middleware.Admin), dummyId, expIn)
	if err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "generate token failed", err)
		return
	}

	ctx = logging.WithLogRole(ctx, string(middleware.Admin))
	out := handler2.DummyLoginOut{
		Token: token,
	}
	if err = json.NewEncoder(w).Encode(out); err != nil {
		handler.RespondWithError(w, ctx, http.StatusInternalServerError, handler2.UNKNOWN, "failed to encode response", err)
		return
	}
}
