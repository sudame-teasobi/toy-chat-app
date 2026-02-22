package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sudame/chat/internal/domain/user"
	"github.com/sudame/chat/internal/service"
)

type mockUserRepo struct {
	saveFunc    func(ctx context.Context, u *user.User) error
	findByIDFunc func(ctx context.Context, id string) (*user.User, error)
}

func (m *mockUserRepo) Save(ctx context.Context, u *user.User) error {
	return m.saveFunc(ctx, u)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	return m.findByIDFunc(ctx, id)
}

func TestCreateUserHandler_Handle(t *testing.T) {
	e := echo.New()

	t.Run("正常系: 有効なリクエストでユーザーを作成してIDを返す", func(t *testing.T) {
		repo := &mockUserRepo{
			saveFunc: func(_ context.Context, _ *user.User) error { return nil },
		}
		svc := service.NewCreateUserService(repo)
		h := NewCreateUserHandler(svc)

		body := `{"name": "Alice"}`
		req := httptest.NewRequest(http.MethodPost, "/create-user", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.Handle(c); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		var resp CreateUserResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if resp.UserID == "" {
			t.Error("expected non-empty userId in response")
		}
	})

	t.Run("異常系: 不正なJSONでは400を返す", func(t *testing.T) {
		repo := &mockUserRepo{
			saveFunc: func(_ context.Context, _ *user.User) error { return nil },
		}
		svc := service.NewCreateUserService(repo)
		h := NewCreateUserHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/create-user", strings.NewReader(`{invalid json`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h.Handle(c) //nolint: errcheck
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("異常系: サービスがエラーを返した場合500を返す", func(t *testing.T) {
		repo := &mockUserRepo{
			saveFunc: func(_ context.Context, _ *user.User) error {
				return errors.New("db error")
			},
		}
		svc := service.NewCreateUserService(repo)
		h := NewCreateUserHandler(svc)

		body := `{"name": "Alice"}`
		req := httptest.NewRequest(http.MethodPost, "/create-user", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h.Handle(c) //nolint: errcheck
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("異常系: 空のユーザー名でサービスエラーを返した場合500を返す", func(t *testing.T) {
		repo := &mockUserRepo{
			saveFunc: func(_ context.Context, _ *user.User) error { return nil },
		}
		svc := service.NewCreateUserService(repo)
		h := NewCreateUserHandler(svc)

		body := `{"name": ""}`
		req := httptest.NewRequest(http.MethodPost, "/create-user", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h.Handle(c) //nolint: errcheck
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}
