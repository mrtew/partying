package room

import (
	"partying/internal/ws"
	"sync"
)

// Room 代表一个语音/聊天房间
// 每个房间有自己的goroutine负责广播消息
type Room struct {
	ID           string
	Name         string
	HostID       string
	MaxCapacity  int
	participants map[string]*Participant // userID -> Participant
	broadcast    chan broadcastMsg       // 广播消息队列
	mu           sync.RWMutex            // 读写锁
}

type broadcastMsg struct {
	msg    ws.OutgoingMessage
	except string // 排除某个用户（发送者自己不收）
}

func NewRoom(id, name, hostID string, maxCapacity int) *Room {
	return &Room{
		ID:           id,
		Name:         name,
		HostID:       hostID,
		MaxCapacity:  maxCapacity,
		participants: make(map[string]*Participant),
		broadcast:    make(chan broadcastMsg, 256),
	}
}

// Run 是房间的主goroutine，负责广播消息给所有人
// 用 go room.Run() 启动
func (r *Room) Run() {
	for msg := range r.broadcast { // 阻塞等待消息
		r.mu.RLock()
		for userID, p := range r.participants {
			if userID == msg.except {
				continue
			}
			p.SendMessage(msg.msg) // 发给每个人
		}
		r.mu.RUnlock()
	}
}

// Broadcast 向房间内所有人广播消息
func (r *Room) Broadcast(msg ws.OutgoingMessage, exceptUserID string) {
	r.broadcast <- broadcastMsg{msg: msg, except: exceptUserID}
}

// AddParticipant 用户加入房间。写操作（加入/离开）用Lock，同一时间只允许一个goroutine
func (r *Room) AddParticipant(p *Participant) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.participants) >= r.MaxCapacity {
		return false // 房间已满
	}
	r.participants[p.UserID] = p
	return true
}

// "读多写少的场景我用RWMutex而不是普通Mutex，读操作可以并发执行，只有写操作才互斥，性能更好。"

// RemoveParticipant 用户离开房间
func (r *Room) RemoveParticipant(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.participants, userID)
}

// GetUsernames 获取房间内所有用户名列表。读操作用RLock，允许多个goroutine同时读
func (r *Room) GetUsernames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.participants))
	for _, p := range r.participants {
		names = append(names, p.Username)
	}
	return names
}

// Count 返回当前房间人数
func (r *Room) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.participants)
}

// IsEmpty 返回房间是否为空
func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.participants) == 0
}
