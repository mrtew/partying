package ws

// 客户端发来的消息类型
const (
	TypeJoinRoom    = "join_room"
	TypeLeaveRoom   = "leave_room"
	TypeChatMessage = "chat_message"
)

// 服务器推送给客户端的消息类型
const (
	TypeRoomJoined = "room_joined" // 成功加入房间
	TypeRoomLeft   = "room_left"   // 成功离开房间
	TypeNewMessage = "new_message" // 新聊天消息
	TypeRoomUsers  = "room_users"  // 房间用户列表更新
	TypeError      = "error"       // 错误通知
	TypeUserJoined = "user_joined" // 有人加入房间
	TypeUserLeft   = "user_left"   // 有人离开房间
)

// IncomingMessage 是客户端发来的消息结构
type IncomingMessage struct {
	Type    string `json:"type"`
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

// OutgoingMessage 是服务器发给客户端的消息结构
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ChatPayload 聊天消息内容
type ChatPayload struct {
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

// UserPayload 用户信息
type UserPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	RoomID   string `json:"room_id"`
}

// RoomUsersPayload 房间用户列表
type RoomUsersPayload struct {
	RoomID string   `json:"room_id"`
	Users  []string `json:"users"`
}

// ErrorPayload 错误信息
type ErrorPayload struct {
	Message string `json:"message"`
}
