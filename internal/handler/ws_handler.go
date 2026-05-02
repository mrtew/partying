package handler

import (
	"net/http"
	"partying/internal/middleware"
	"partying/internal/room"
	"partying/pkg/response"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源（开发环境用，生产环境要限制）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	manager *room.Manager
}

func NewWSHandler(manager *room.Manager) *WSHandler {
	return &WSHandler{manager: manager}
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	username := middleware.GetUsername(r)

	if userID == "" {
		response.Unauthorized(w, "unauthorized")
		return
	}

	// HTTP升级为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		response.InternalError(w, "failed to upgrade connection")
		return
	}

	// 创建参与者
	participant := room.NewParticipant(userID, username, conn, h.manager)
	h.manager.Register(participant)

	// 每个连接启动两个goroutine：一读一写
	go participant.ReadPump()  // 专门"听"这个用户说什么
	go participant.WritePump() // 专门"发"消息给这个用户
	/*
		用户A连接 → go ReadPump(A) + go WritePump(A)
		用户B连接 → go ReadPump(B) + go WritePump(B)
		用户C连接 → go ReadPump(C) + go WritePump(C)
		3个用户 = 6个goroutine同时跑，互不阻塞
		1000个用户 = 2000个goroutine，Go轻松处理
		"每个WebSocket连接我用两个goroutine分离读写，读写互不阻塞，这是Go的标准WebSocket并发模式。"
	*/
}
