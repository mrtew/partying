package model

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // "-" 表示不会出现在JSON响应里
	CreatedAt time.Time `json:"created_at"`
}
