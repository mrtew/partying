package model

import "time"

type Room struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	HostID      string    `json:"host_id"`
	MaxCapacity int       `json:"max_capacity"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoomInfo 是返回给客户端的房间信息（包含当前人数）
// type RoomInfo struct {
// 	Room
// 	CurrentUsers int `json:"current_users"`
// }
