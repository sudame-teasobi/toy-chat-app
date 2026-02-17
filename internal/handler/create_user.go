package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudame/chat/internal/service"
)

type CreateUserRequest struct {
	Name string `json:"name"`
}

type CreateUserResponse struct {
	UserID string `json:"userId"`
}

type CreateUserHandler struct {
	service *service.CreateUserService
}

func NewCreateUserHandler(s *service.CreateUserService) *CreateUserHandler {
	return &CreateUserHandler{
		service: s,
	}
}

func (h *CreateUserHandler) Handle(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] /create-user: failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	log.Printf("[INFO] /create-user: request received: name=%q", req.Name)

	userId, err := h.service.Exec(c.Request().Context(), req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("internal server errror: %s", err.Error())})
	}

	return c.JSON(http.StatusOK, CreateUserResponse{UserID: *userId})
}
