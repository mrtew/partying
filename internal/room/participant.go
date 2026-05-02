package room

import (
	"encoding/json"
	"log"
	"partying/internal/ws"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

// Participant 代表一个通过WebSocket连接到服务器的用户
type Participant struct {
	UserID   string
	Username string
	RoomID   string // 当前所在房间，空字符串表示不在任何房间
	conn     *websocket.Conn
	send     chan []byte // 该用户的消息发送队列
	manager  *Manager
}

func NewParticipant(userID, username string, conn *websocket.Conn, manager *Manager) *Participant {
	return &Participant{
		UserID:   userID,
		Username: username,
		conn:     conn,
		send:     make(chan []byte, 256),
		manager:  manager,
	}
}

// SendMessage 把消息放进该用户的发送队列
func (p *Participant) SendMessage(msg ws.OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case p.send <- data:
	default:
		// 队列满了，说明客户端太慢，关掉连接
		close(p.send)
	}
}

// ReadPump 专门负责"读"：持续读取客户端发来的消息
// 每个连接一个goroutine
func (p *Participant) ReadPump() {
	defer func() {
		// 用户断线时，通知manager清理
		p.manager.unregister <- p
	}()

	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error {
		p.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := p.conn.ReadMessage()
		if err != nil {
			// 连接断开（用户关闭浏览器等），正常退出
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", p.Username, err)
			}
			break
		}

		var msg ws.IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			p.SendMessage(ws.OutgoingMessage{
				Type:    ws.TypeError,
				Payload: ws.ErrorPayload{Message: "invalid message format"},
			})
			continue
		}

		// 把收到的消息交给manager处理
		p.manager.handleMessage(p, msg)
	}
}

// WritePump 专门负责"写"：持续把发送队列里的消息写给客户端
// 每个连接一个goroutine
func (p *Participant) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()

	for {
		select {
		case data, ok := <-p.send:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// send channel被关闭，断开连接
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := p.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			// 定时发ping，确认客户端还活着
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
