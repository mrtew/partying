package service

import (
	"errors"
	"partying/internal/model"
	"partying/internal/repository"
	"time"

	"github.com/google/uuid"
)

type RoomService struct {
	roomRepo   *repository.RoomRepository
	onlineRepo *repository.OnlineRepository
}

func NewRoomService(roomRepo *repository.RoomRepository, onlineRepo *repository.OnlineRepository) *RoomService {
	return &RoomService{
		roomRepo:   roomRepo,
		onlineRepo: onlineRepo,
	}
}

type CreateRoomInput struct {
	Name        string `json:"name"`
	MaxCapacity int    `json:"max_capacity"`
}

// RoomInfo 是返回给客户端的房间信息，额外包含当前实时人数
type RoomInfo struct {
	model.Room
	CurrentUsers int `json:"current_users"`
}

func (s *RoomService) CreateRoom(input CreateRoomInput, hostID string) (*model.Room, error) {
	if input.Name == "" {
		return nil, errors.New("room name is required")
	}
	if input.MaxCapacity <= 0 {
		input.MaxCapacity = 10
	}
	if input.MaxCapacity > 50 {
		return nil, errors.New("max capacity cannot exceed 50")
	}

	room := &model.Room{
		ID:          uuid.New().String(),
		Name:        input.Name,
		HostID:      hostID,
		MaxCapacity: input.MaxCapacity,
		CreatedAt:   time.Now(),
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) ListRooms() ([]*RoomInfo, error) {
	rooms, err := s.roomRepo.FindAll()
	if err != nil {
		return nil, err
	}

	result := make([]*RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		count, _ := s.onlineRepo.GetRoomCount(room.ID)
		result = append(result, &RoomInfo{
			Room:         *room,
			CurrentUsers: int(count),
		})
	}
	return result, nil
}

func (s *RoomService) GetRoom(id string) (*RoomInfo, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	count, _ := s.onlineRepo.GetRoomCount(id)
	return &RoomInfo{Room: *room, CurrentUsers: int(count)}, nil
}
