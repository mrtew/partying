package main

import (
	"fmt"
	"log"
	"net/http"
	"partying/internal/config"
	"partying/internal/database"
	"partying/internal/handler"
	"partying/internal/repository"
	"partying/internal/room"
	"partying/internal/service"
)

func main() {
	// 1. 加载配置
	cfg := config.New()

	// 2. 连接 MySQL
	db, err := database.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("❌ MySQL error: %v", err)
	}
	defer db.Close()

	// 3. 连接 Redis
	rdb, err := database.NewRedis(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("❌ Redis error: %v", err)
	}

	// Repository层
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	onlineRepo := repository.NewOnlineRepository(rdb)

	// Service层
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	roomService := service.NewRoomService(roomRepo, onlineRepo)

	// Room Manager（WebSocket核心）
	manager := room.NewManager(roomRepo, onlineRepo)

	// Handler层
	authHandler := handler.NewAuthHandler(authService)
	roomHandler := handler.NewRoomHandler(roomService)
	wsHandler := handler.NewWSHandler(manager)

	// 路由
	router := handler.NewRouter(authHandler, roomHandler, wsHandler, authService, manager)

	fmt.Println("🎉 Partying App server is running!")
	fmt.Printf("👉 Health:    GET  http://localhost%s/health\n", cfg.Port)
	fmt.Printf("👉 Register:  POST http://localhost%s/api/register\n", cfg.Port)
	fmt.Printf("👉 Login:     POST http://localhost%s/api/login\n", cfg.Port)
	fmt.Printf("👉 Rooms:     GET  http://localhost%s/api/rooms\n", cfg.Port)
	fmt.Printf("👉 WebSocket: WS   ws://localhost%s/ws\n", cfg.Port)

	if err := http.ListenAndServe(cfg.Port, router); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
