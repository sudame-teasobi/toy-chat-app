package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudame/chat/internal/applicationservice"
	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/domain/user"
)

type CreateChatRoomRequest struct {
	Name      string `json:"name"`
	CreatorID string `json:"creator_id"`
}

type MemberResponse struct {
	UserID string `json:"user_id"`
}

type CreateChatRoomResponse struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Members []MemberResponse `json:"members"`
}

type CreateChatRoomHandler struct {
	usecase *applicationservice.CreateChatRoomUsecase
}

func NewCreateChatRoomHandler(chatRoomRepo chatroom.Repository, userRepo user.Repository) *CreateChatRoomHandler {
	return &CreateChatRoomHandler{
		usecase: applicationservice.NewCreateChatRoomUsecase(chatRoomRepo, userRepo),
	}
}

func (h *CreateChatRoomHandler) Handle(c echo.Context) error {
	var req CreateChatRoomRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] /create-chat-room: failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	log.Printf("[INFO] /create-chat-room: request received: name=%q, creator_id=%q", req.Name, req.CreatorID)

	input := applicationservice.CreateChatRoomInput{
		Name:      req.Name,
		CreatorID: req.CreatorID,
	}

	output, err := h.usecase.Execute(c.Request().Context(), input)
	if err != nil {
		log.Printf("[ERROR] /create-chat-room: usecase execution failed: %v", err)
		switch err {
		case chatroom.ErrEmptyName:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		case user.ErrNotFound:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "creator not found"})
		default:
			log.Printf("[ERROR] /create-chat-room: unexpected error: %+v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}

	return c.JSON(http.StatusCreated, CreateChatRoomResponse{
		ID:   output.ChatRoom.ID(),
		Name: output.ChatRoom.Name(),
	})
}
