package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// OnlineRepository 用Redis追踪实时状态（谁在哪个房间）
type OnlineRepository struct {
	redis *redis.Client
}

func NewOnlineRepository(redis *redis.Client) *OnlineRepository {
	return &OnlineRepository{redis: redis}
}

// Redis key格式: "room:abc123:participants"
func roomKey(roomID string) string {
	return fmt.Sprintf("room:%s:participants", roomID)
}

// AddToRoom 用户加入房间（Redis SET集合）
func (r *OnlineRepository) AddToRoom(roomID, userID string) error {
	return r.redis.SAdd(context.Background(), roomKey(roomID), userID).Err()
}

// RemoveFromRoom 用户离开房间
func (r *OnlineRepository) RemoveFromRoom(roomID, userID string) error {
	return r.redis.SRem(context.Background(), roomKey(roomID), userID).Err()
}

// GetRoomParticipants 获取房间内所有用户ID
func (r *OnlineRepository) GetRoomParticipants(roomID string) ([]string, error) {
	return r.redis.SMembers(context.Background(), roomKey(roomID)).Result()
}

// GetRoomCount 获取房间当前人数
func (r *OnlineRepository) GetRoomCount(roomID string) (int64, error) {
	return r.redis.SCard(context.Background(), roomKey(roomID)).Result()
}
