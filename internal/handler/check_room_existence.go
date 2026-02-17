package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudame/chat/internal/service"
)

type CheckRoomExistenceRequest struct {
	RoomID string `json:"room_id"`
}

type CheckRoomExistenceResponse struct {
	Existence bool `json:"existence"`
}

type CheckRoomExistenceHandler struct {
	service *service.CheckRoomExistenceService
}

func NewCheckRoomExistenceHandler(s *service.CheckRoomExistenceService) *CheckRoomExistenceHandler {
	return &CheckRoomExistenceHandler{
		service: s,
	}
}

func (h *CheckRoomExistenceHandler) Handle(c echo.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("internal server error: %s", r)})
		}
	}()

	var req CheckRoomExistenceRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] /check_room_existence: failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	log.Printf("[INFO] /check_room_existence: request received: room_id=%q", req.RoomID)

	result := h.service.Exec(c.Request().Context(), req.RoomID)
	return c.JSON(http.StatusOK, CheckRoomExistenceResponse{Existence: result.Existence})
}
