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
	"github.com/sudame/chat/internal/service"
)

func TestCheckRoomExistenceHandler_Handle(t *testing.T) {
	e := echo.New()

	t.Run("正常系: ルームが存在する場合 existence=true を返す", func(t *testing.T) {
		roomRepo := &mockRoomRepo{
			findByIDFunc: func(_ context.Context, _ string) (*room.Room, error) {
				return room.ReconstructRoom("chat-room:01", "テストルーム"), nil
			},
		}
		svc := service.NewCheckRoomExistenceService(roomRepo)
		h := NewCheckRoomExistenceHandler(svc)

		body := `{"room_id": "chat-room:01"}`
		req := httptest.NewRequest(http.MethodPost, "/check-room-existence", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.Handle(c); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		var resp CheckRoomExistenceResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if !resp.Existence {
			t.Error("expected existence=true")
		}
	})

	t.Run("正常系: ルームが存在しない場合 existence=false を返す", func(t *testing.T) {
		roomRepo := &mockRoomRepo{
			findByIDFunc: func(_ context.Context, _ string) (*room.Room, error) {
				return nil, room.ErrNotFound
			},
		}
		svc := service.NewCheckRoomExistenceService(roomRepo)
		h := NewCheckRoomExistenceHandler(svc)

		body := `{"room_id": "chat-room:nonexistent"}`
		req := httptest.NewRequest(http.MethodPost, "/check-room-existence", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.Handle(c); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		var resp CheckRoomExistenceResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if resp.Existence {
			t.Error("expected existence=false")
		}
	})

	t.Run("異常系: リポジトリが予期しないエラーを返す場合500を返す", func(t *testing.T) {
		roomRepo := &mockRoomRepo{
			findByIDFunc: func(_ context.Context, _ string) (*room.Room, error) {
				return nil, errors.New("db error")
			},
		}
		svc := service.NewCheckRoomExistenceService(roomRepo)
		h := NewCheckRoomExistenceHandler(svc)

		body := `{"room_id": "chat-room:01"}`
		req := httptest.NewRequest(http.MethodPost, "/check-room-existence", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h.Handle(c) //nolint: errcheck
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("異常系: 不正なJSONでは400を返す", func(t *testing.T) {
		roomRepo := &mockRoomRepo{
			findByIDFunc: func(_ context.Context, _ string) (*room.Room, error) {
				return nil, nil
			},
		}
		svc := service.NewCheckRoomExistenceService(roomRepo)
		h := NewCheckRoomExistenceHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/check-room-existence", strings.NewReader(`{invalid json`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h.Handle(c) //nolint: errcheck
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}
