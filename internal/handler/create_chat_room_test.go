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
	"github.com/sudame/chat/internal/domain/room"
	"github.com/sudame/chat/internal/domain/user"
	"github.com/sudame/chat/internal/service"
)

type mockRoomRepo struct {
	saveFunc    func(ctx context.Context, r *room.Room) error
	findByIDFunc func(ctx context.Context, id string) (*room.Room, error)
}

func (m *mockRoomRepo) Save(ctx context.Context, r *room.Room) error {
	return m.saveFunc(ctx, r)
}

func (m *mockRoomRepo) FindByID(ctx context.Context, id string) (*room.Room, error) {
	return m.findByIDFunc(ctx, id)
}

func TestCreateRoomHandler_Handle(t *testing.T) {
	e := echo.New()

	t.Run("正常系: 有効なリクエストでルームを作成してIDを返す", func(t *testing.T) {
		userRepo := &mockUserRepo{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomRepo := &mockRoomRepo{
			saveFunc: func(_ context.Context, _ *room.Room) error { return nil },
		}
		svc := service.NewCreateRoomService(userRepo, roomRepo)
		h := NewCreateRoomHandler(svc)

		body := `{"name": "テストルーム", "creator_id": "user:01"}`
		req := httptest.NewRequest(http.MethodPost, "/create-chat-room", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.Handle(c); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		var resp CreateRoomResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if resp.RoomID == "" {
			t.Error("expected non-empty room_id in response")
		}
	})

	t.Run("異常系: 不正なJSONでは400を返す", func(t *testing.T) {
		userRepo := &mockUserRepo{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomRepo := &mockRoomRepo{
			saveFunc: func(_ context.Context, _ *room.Room) error { return nil },
		}
		svc := service.NewCreateRoomService(userRepo, roomRepo)
		h := NewCreateRoomHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/create-chat-room", strings.NewReader(`{invalid json`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h.Handle(c) //nolint: errcheck
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("異常系: 存在しないユーザーIDでは500を返す", func(t *testing.T) {
		userRepo := &mockUserRepo{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return nil, user.ErrNotFound
			},
		}
		roomRepo := &mockRoomRepo{
			saveFunc: func(_ context.Context, _ *room.Room) error { return nil },
		}
		svc := service.NewCreateRoomService(userRepo, roomRepo)
		h := NewCreateRoomHandler(svc)

		body := `{"name": "テストルーム", "creator_id": "nonexistent"}`
		req := httptest.NewRequest(http.MethodPost, "/create-chat-room", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h.Handle(c) //nolint: errcheck
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("異常系: リポジトリのSaveエラーでは500を返す", func(t *testing.T) {
		userRepo := &mockUserRepo{
			findByIDFunc: func(_ context.Context, _ string) (*user.User, error) {
				return user.ReconstructUser("user:01", "Alice"), nil
			},
		}
		roomRepo := &mockRoomRepo{
			saveFunc: func(_ context.Context, _ *room.Room) error {
				return errors.New("db error")
			},
		}
		svc := service.NewCreateRoomService(userRepo, roomRepo)
		h := NewCreateRoomHandler(svc)

		body := `{"name": "テストルーム", "creator_id": "user:01"}`
		req := httptest.NewRequest(http.MethodPost, "/create-chat-room", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h.Handle(c) //nolint: errcheck
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}
