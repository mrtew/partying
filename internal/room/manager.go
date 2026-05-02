package room

import (
	"log"
	"partying/internal/repository"
	"partying/internal/ws"
	"sync"
)

// Manager 管理所有活跃的房间和WebSocket参与者
type Manager struct {
	rooms   map[string]*Room        // roomID -> Room
	clients map[string]*Participant // userID -> Participant
	mu      sync.RWMutex

	// Channel就像一条传送带
	register   chan *Participant // 新用户连接，有人连接，放进这条传送带
	unregister chan *Participant // 用户断开连接，有人断线，放进这条传送带

	roomRepo   *repository.RoomRepository
	onlineRepo *repository.OnlineRepository
}

func NewManager(roomRepo *repository.RoomRepository, onlineRepo *repository.OnlineRepository) *Manager {
	m := &Manager{
		rooms:      make(map[string]*Room),
		clients:    make(map[string]*Participant),
		register:   make(chan *Participant, 16),
		unregister: make(chan *Participant, 16),
		roomRepo:   roomRepo,
		onlineRepo: onlineRepo,
	}
	go m.run()
	return m
}

// run 是Manager的主goroutine，处理用户注册和注销，manager的goroutine专门从传送带取东西处理
func (m *Manager) run() {
	for {
		select { // select = 同时监听多条传送带
		case p := <-m.register: // 处理新连接
			m.mu.Lock()
			m.clients[p.UserID] = p
			m.mu.Unlock()
			log.Printf("✅ User connected: %s", p.Username)

		case p := <-m.unregister: // 处理断线
			m.mu.Lock()
			if _, exists := m.clients[p.UserID]; exists {
				delete(m.clients, p.UserID)
				// 如果用户在房间里，自动离开
				if p.RoomID != "" {
					m.leaveRoom(p)
				}
				close(p.send)
			}
			m.mu.Unlock()
			log.Printf("❌ User disconnected: %s", p.Username)
		}
	}
}

/*
❌ 危险做法（没有channel）：
goroutine A 和 goroutine B 同时写同一个map
→ Race condition → 数据损坏 → 程序崩溃
✅ 安全做法（用channel）：
goroutine A 把操作放进channel
goroutine B 把操作放进channel
manager goroutine 一个一个取出来处理
→ 永远只有一个goroutine操作map → 安全
 "我用Channel来通信而不是共享内存，这是Go的核心设计理念：Don't communicate by sharing memory, share memory by communicating。"
*/

// Register 注册新的WebSocket连接
func (m *Manager) Register(p *Participant) {
	m.register <- p
}

// handleMessage 处理客户端发来的各种消息
func (m *Manager) handleMessage(p *Participant, msg ws.IncomingMessage) {
	switch msg.Type {
	case ws.TypeJoinRoom:
		m.mu.Lock()
		m.joinRoom(p, msg.RoomID)
		m.mu.Unlock()

	case ws.TypeLeaveRoom:
		m.mu.Lock()
		m.leaveRoom(p)
		m.mu.Unlock()

	case ws.TypeChatMessage:
		m.handleChat(p, msg)

	default:
		p.SendMessage(ws.OutgoingMessage{
			Type:    ws.TypeError,
			Payload: ws.ErrorPayload{Message: "unknown message type"},
		})
	}
}

// joinRoom 处理用户加入房间的逻辑（调用前需要持有mu锁）
func (m *Manager) joinRoom(p *Participant, roomID string) {
	// 如果用户已经在某个房间，先离开
	if p.RoomID != "" {
		m.leaveRoom(p)
	}

	// 从DB确认房间存在
	roomModel, err := m.roomRepo.FindByID(roomID)
	if err != nil {
		p.SendMessage(ws.OutgoingMessage{
			Type:    ws.TypeError,
			Payload: ws.ErrorPayload{Message: "room not found"},
		})
		return
	}

	// 获取或创建内存中的Room实例
	room, exists := m.rooms[roomID]
	if !exists {
		room = NewRoom(roomModel.ID, roomModel.Name, roomModel.HostID, roomModel.MaxCapacity)
		m.rooms[roomID] = room
		go room.Run() // 启动房间广播goroutine，这个房间有自己的goroutine处理广播
		/*
			房间A的goroutine → 只处理房间A的广播
			房间B的goroutine → 只处理房间B的广播
			房间C的goroutine → 只处理房间C的广播
			完全独立，互不影响
			"每个房间有独立goroutine负责消息广播，房间之间完全隔离，不会互相阻塞。"
		*/
	}

	// 加入房间
	if ok := room.AddParticipant(p); !ok {
		p.SendMessage(ws.OutgoingMessage{
			Type:    ws.TypeError,
			Payload: ws.ErrorPayload{Message: "room is full"},
		})
		return
	}

	p.RoomID = roomID
	m.onlineRepo.AddToRoom(roomID, p.UserID)

	// 通知该用户：成功加入
	p.SendMessage(ws.OutgoingMessage{
		Type: ws.TypeRoomJoined,
		Payload: ws.UserPayload{
			UserID:   p.UserID,
			Username: p.Username,
			RoomID:   roomID,
		},
	})

	// 通知房间其他人：有新人加入
	room.Broadcast(ws.OutgoingMessage{
		Type: ws.TypeUserJoined,
		Payload: ws.UserPayload{
			UserID:   p.UserID,
			Username: p.Username,
			RoomID:   roomID,
		},
	}, p.UserID)

	// 广播更新后的用户列表
	room.Broadcast(ws.OutgoingMessage{
		Type: ws.TypeRoomUsers,
		Payload: ws.RoomUsersPayload{
			RoomID: roomID,
			Users:  room.GetUsernames(),
		},
	}, "")

	log.Printf("👥 %s joined room %s (%d/%d)", p.Username, roomModel.Name, room.Count(), roomModel.MaxCapacity)
}

// leaveRoom 处理用户离开房间（调用前需要持有mu锁）
func (m *Manager) leaveRoom(p *Participant) {
	if p.RoomID == "" {
		return
	}

	room, exists := m.rooms[p.RoomID]
	if !exists {
		return
	}

	room.RemoveParticipant(p.UserID)
	m.onlineRepo.RemoveFromRoom(p.RoomID, p.UserID)

	// 通知该用户：已离开
	p.SendMessage(ws.OutgoingMessage{
		Type: ws.TypeRoomLeft,
		Payload: ws.UserPayload{
			UserID:   p.UserID,
			Username: p.Username,
			RoomID:   p.RoomID,
		},
	})

	// 通知房间其他人
	room.Broadcast(ws.OutgoingMessage{
		Type: ws.TypeUserLeft,
		Payload: ws.UserPayload{
			UserID:   p.UserID,
			Username: p.Username,
			RoomID:   p.RoomID,
		},
	}, p.UserID)

	// 更新用户列表
	if !room.IsEmpty() {
		room.Broadcast(ws.OutgoingMessage{
			Type: ws.TypeRoomUsers,
			Payload: ws.RoomUsersPayload{
				RoomID: p.RoomID,
				Users:  room.GetUsernames(),
			},
		}, "")
	} else {
		// 房间空了，清理掉
		delete(m.rooms, p.RoomID)
	}

	log.Printf("👋 %s left room %s", p.Username, p.RoomID)
	p.RoomID = ""
}

// handleChat 处理聊天消息
func (m *Manager) handleChat(p *Participant, msg ws.IncomingMessage) {
	if p.RoomID == "" {
		p.SendMessage(ws.OutgoingMessage{
			Type:    ws.TypeError,
			Payload: ws.ErrorPayload{Message: "you are not in a room"},
		})
		return
	}

	m.mu.RLock()
	room, exists := m.rooms[p.RoomID]
	m.mu.RUnlock()

	if !exists {
		return
	}

	// 广播聊天消息给房间所有人（包括发送者自己）
	room.Broadcast(ws.OutgoingMessage{
		Type: ws.TypeNewMessage,
		Payload: ws.ChatPayload{
			RoomID:   p.RoomID,
			UserID:   p.UserID,
			Username: p.Username,
			Content:  msg.Content,
		},
	}, "")
}
