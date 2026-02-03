package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudame/chat/internal/service"
)

type CreateRoomRequest struct {
	Name      string `json:"name"`
	CreatorID string `json:"creator_id"`
}

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
}

type CreateRoomHandler struct {
	service service.CreateRoomService
}

func NewCreateRoomHandler(createRoomService service.CreateRoomService) *CreateRoomHandler {
	return &CreateRoomHandler{}
}

func (h *CreateRoomHandler) Handle(c echo.Context) error {
	var req CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] /create-chat-room: failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	log.Printf("[INFO] /create-chat-room: request received: name=%q, creator_id=%q", req.Name, req.CreatorID)

	roomID, err := h.service.Exec(c.Request().Context(), req.Name, req.CreatorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("internal server error: %s", err.Error())})
	}

	return c.JSON(http.StatusOK, CreateRoomResponse{RoomID: *roomID})
}
