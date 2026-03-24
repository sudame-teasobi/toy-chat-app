package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudame/chat/internal/service"
)

type CheckMembershipExistenceRequest struct {
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
}

type CheckMembershipExistenceResponse struct {
	Existence bool `json:"existence"`
}

type CheckMembershipExistenceHandler struct {
	service *service.CheckMembershipExistenceService
}

func NewCheckMembershipExistenceHandler(s *service.CheckMembershipExistenceService) *CheckMembershipExistenceHandler {
	return &CheckMembershipExistenceHandler{
		service: s,
	}
}

func (h *CheckMembershipExistenceHandler) Handle(c echo.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("internal server error: %s", r)})
		}
	}()

	var req CheckMembershipExistenceRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] /check_room_existence: failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	result, err := h.service.Exec(c.Request().Context(), req.UserID, req.RoomID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, CheckMembershipExistenceResponse{Existence: result.Existence})
}
