package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudame/chat/internal/applicationservice"
	"github.com/sudame/chat/internal/domain/user"
)

type CreateUserRequest struct {
	Name string `json:"name"`
}

type CreateUserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateUserHandler struct {
	usecase *applicationservice.CreateUserUsecase
}

func NewCreateUserHandler(userRepo user.Repository) *CreateUserHandler {
	return &CreateUserHandler{
		usecase: applicationservice.NewCreateUserUsecase(userRepo),
	}
}

func (h *CreateUserHandler) Handle(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] /create-user: failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	log.Printf("[INFO] /create-user: request received: name=%q", req.Name)

	input := applicationservice.CreateUserInput{
		Name: req.Name,
	}

	output, err := h.usecase.Execute(c.Request().Context(), input)
	if err != nil {
		log.Printf("[ERROR] /create-user: usecase execution failed: %v", err)
		switch err {
		case user.ErrEmptyName:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			log.Printf("[ERROR] /create-user: unexpected error: %+v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}

	return c.JSON(http.StatusCreated, CreateUserResponse{
		ID:   output.User.ID(),
		Name: output.User.Name(),
	})
}
