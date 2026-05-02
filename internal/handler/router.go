package handler

import (
	"net/http"
	"partying/internal/middleware"
	"partying/internal/room"
	"partying/internal/service"
	"partying/pkg/response"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	authHandler *AuthHandler,
	roomHandler *RoomHandler,
	wsHandler *WSHandler,
	authService *service.AuthService,
	manager *room.Manager,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// 直接serve web/index.html
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"status": "ok", "service": "partying"})
	})

	// 公开路由
	r.Post("/api/register", authHandler.Register)
	r.Post("/api/login", authHandler.Login)

	// 需要JWT的路由
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService))
		r.Get("/api/rooms", roomHandler.ListRooms)
		r.Post("/api/rooms", roomHandler.CreateRoom)
		r.Get("/api/rooms/{id}", roomHandler.GetRoom)
		// WebSocket端点
		r.Get("/ws", wsHandler.ServeWS)
	})

	return r
}
