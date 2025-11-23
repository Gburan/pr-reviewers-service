package dummy_login_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handler2 "pr-reviewers-service/internal/generated/api/v1/handler"
	"pr-reviewers-service/internal/handler/dummy_login"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDummyLogin_Success(t *testing.T) {
	validate := validator.New()
	handler := dummy_login.New("secret123", validate)

	req := httptest.NewRequest("POST", "/dummyLogin", nil)
	w := httptest.NewRecorder()

	handler.DummyLogin(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp handler2.DummyLoginOut
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
}
