package handler

import (
	"encoding/json"
	"net/http"
	"partying/internal/middleware"
	"partying/internal/service"
	"partying/pkg/response"

	"github.com/go-chi/chi/v5"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomService.ListRooms()
	if err != nil {
		response.InternalError(w, "failed to fetch rooms")
		return
	}
	response.OK(w, rooms)
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var input service.CreateRoomInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	room, err := h.roomService.CreateRoom(input, userID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, room)
}

func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	room, err := h.roomService.GetRoom(id)
	if err != nil {
		response.BadRequest(w, "room not found")
		return
	}
	response.OK(w, room)
}
