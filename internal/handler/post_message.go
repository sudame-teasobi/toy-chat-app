package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudame/chat/internal/domain/message"
	"github.com/sudame/chat/internal/service"
)

type PostMessageRequest struct {
	AuthorUserID string `json:"author_user_id"`
	RoomID       string `json:"room_id"`
	Body         string `json:"body"`
}

type PostMessageResponse struct {
	MessageID string `json:"message_id"`
}

type PostMessageHandler struct {
	service *service.PostMessageService
}

func NewPostMessageHandler(s *service.PostMessageService) *PostMessageHandler {
	return &PostMessageHandler{
		service: s,
	}
}

func (h *PostMessageHandler) Handle(c echo.Context) error {
	var req PostMessageRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] /post-message: failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	messageID, err := h.service.Exec(c.Request().Context(), req.AuthorUserID, req.RoomID, req.Body)
	if err != nil {
		if errors.Is(err, message.ErrForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": fmt.Sprintf("not a member of this room: %s", err.Error())})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("internal server errror: %s", err.Error())})
	}

	return c.JSON(http.StatusOK, PostMessageResponse{MessageID: *messageID})
}
