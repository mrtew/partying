# partying

/////////////////////////////////////////////////////

Directory:

partying/
│
├── cmd/
│   └── server/
│       └── main.go                  # 程序入口，只负责启动
│
├── internal/                        # 核心业务逻辑
│   ├── config/
│   │   └── config.go                # 所有配置（端口、JWT secret等）
│   │
│   ├── model/                       # 纯数据结构，不含逻辑
│   │   ├── user.go                  # 用户模型
│   │   └── room.go                  # 房间模型
│   │
│   ├── repository/                  # 数据存取层（读写内存/数据库）
│   │   ├── user_repository.go
│   │   └── room_repository.go
│   │
│   ├── service/                     # 业务逻辑层
│   │   ├── auth_service.go          # 注册/登录逻辑
│   │   └── room_service.go          # 房间业务逻辑
│   │
│   ├── handler/                     # HTTP处理层（只负责解析请求/返回响应）
│   │   ├── auth_handler.go          # 注册/登录接口
│   │   ├── room_handler.go
│   │   └── ws_handler.go            # WebSocket连接处理
│   │
│   ├── middleware/
│   │   └── auth.go                  # JWT验证中间件
│   │
│   ├── room/                        # 并发核心：房间goroutine逻辑，房间核心逻辑 ⭐最重要
│   │   ├── participant.go           # 参与者（WebSocket连接）
│   │   ├── room.go                  # 单个房间（goroutine + channel）
│   │   └── manager.go               # 所有房间的管理器
│   │
│   └── ws/
│       └── message.go               # WebSocket消息类型定义
│
├── pkg/
│   └── response/
│       └── response.go              # 统一HTTP响应格式
│
├── web/
│   └── index.html                   # 极简前端demo
│
├── go.mod
├── go.sum
└── README.md


语音部分我们保留WebRTC信令的后端代码（面试时可以讲解架构），但demo时主要展示实时聊天+多用户同步，这已经足够展示Go并发的核心能力了。

面试时的demo剧本
"我来演示一下，我开三个标签模拟三个用户..."
创建房间
其他用户加入
实时聊天同步
翻开代码讲："这里用goroutine+channel实现广播，保证并发安全..."
这样可视化效果强，又能引出Go并发的技术深度讨论
///////////////////////////////////////////////////////////////////////////

System Architecture:

客户端 (Postman / 前端)
        │
        ├── HTTP REST ──────────► api/router.go
        │                              │
        │                    ┌─────────┴──────────┐
        │                 auth/              room list
        │               登录/注册            创建/列表
        │
        └── WebSocket ─────────► ws/handler.go
                                       │
                              internal/room/manager.go
                                       │
                          ┌────────────┼────────────┐
                       Room A        Room B        Room C
                    (goroutine)   (goroutine)   (goroutine)
                    [chan msg]    [chan msg]    [chan msg]
                      │
              ┌───────┴────────┐
          User 1           User 2
        (WebSocket)      (WebSocket)
         发SDP offer  ──►  收SDP answer
              ◄──── ICE candidate ────►
                  （完成WebRTC握手）
                  🎙️ 音频直接P2P传输

/////////////////////////////////////////////////////
阶段内容覆盖Go考点Step 1项目初始化 + 模型定义struct, go modStep 2统一响应格式 + 路由interface, httpStep 3注册/登录 + JWTerror handling, packageStep 4房间管理器map, mutex, struct methodStep 5房间goroutine + channel⭐ goroutine, channel, selectStep 6WebSocket连接goroutine, for-select patternStep 7WebRTC信令JSON, interface, switchStep 8整合测试 + README完整项目
//////////////////////////////////////////////////////

Functions:

POST /api/register        # 注册
POST /api/login           # 登录，返回JWT token

GET  /api/rooms           # 获取所有房间列表
POST /api/rooms           # 创建新房间

WS   /ws?token=xxx        # WebSocket连接（需要JWT）

── WebSocket消息类型 ──
join_room                 # 加入房间
leave_room                # 离开房间
chat_message              # 文字聊天
voice_offer               # WebRTC: 发起语音连接
voice_answer              # WebRTC: 回应语音连接
ice_candidate             # WebRTC: 网络穿透
room_users                # 服务器推送：当前房间用户列表

/////////////////////////////////////////////////////

HTTP请求 → Handler（接收请求）
              ↓
           Service（业务逻辑）
              ↓
        Repository（数据存取）

/////////////////////////////////////////////////////

举例：用户调用 POST /api/rooms 创建房间

auth_handler.go   → 我只负责：解析JSON、验证参数、调Service、返回响应
      ↓
room_service.go   → 我只负责：检查房间名是否重复、设置默认值等业务规则
      ↓
room_repository.go → 我只负责：把数据存进内存map（或以后换成数据库）

////////////////////////////////////////////

1. 注册
POST http://localhost:8080/api/register
Body: {"username": "alice", "password": "123456"}
2. 登录（拿token）
POST http://localhost:8080/api/login
Body: {"username": "alice", "password": "123456"}
3. 创建房间（要带token）
POST http://localhost:8080/api/rooms
Header: Authorization: Bearer <你的token>
Body: {"name": "Alice's Party", "max_capacity": 10}
4. 查看所有房间
GET http://localhost:8080/api/rooms
Header: Authorization: Bearer <你的token>

/////////////////////////

MySQL = 硬盘数据库，数据永久保存，但相对慢
         → 存用户账号、房间信息

Redis = 内存数据库，数据存在RAM，超级快，但重启会丢失
         → 存"谁在线"、"房间里现在有几人"这种实时数据

用户注册/登录 → MySQL（要永久保存）
房间实时人数 → Redis（只需要当下，快就好）

//////////////////////

普通程序 = 一个服务员，同时只能服务一个客人
           客人A点餐中... 客人B必须等... 客人C也等...

Goroutine = 餐厅开了很多服务员
           go 服务客人A    ← 一个goroutine
           go 服务客人B    ← 另一个goroutine  
           go 服务客人C    ← 又一个goroutine
           全部同时进行，互不干扰

////////////////////////////////////////////////

用户A的ReadPump goroutine
    │ 收到消息
    ▼
Manager的handleMessage()
    │ 处理业务逻辑
    ▼
Room A的broadcast channel ←── 传送带
    │
    ▼
Room A的Run() goroutine
    │ 取出消息，广播给所有人
    ├──► 用户A的WritePump goroutine → 发给用户A
    ├──► 用户B的WritePump goroutine → 发给用户B
    └──► 用户C的WritePump goroutine → 发给用户C

"每个用户连接有独立的ReadPump和WritePump goroutine，所以1000个连接就是2000个goroutine同时跑，Go的goroutine非常轻量，每个只占几KB内存，2000个完全没问题。消息广播通过buffered channel传递，Room的Run goroutine从channel取消息再分发，避免了直接操作共享内存的race condition。如果房间满了（超过MaxCapacity），AddParticipant会返回false，用户会收到error消息。"

//////////////////////////////////////

"这是我用Go做的简化版Partying语音App后端。采用三层架构：Handler、Service、Repository。并发核心是每个WebSocket连接有独立的读写goroutine，每个房间有独立的广播goroutine通过channel分发消息，用RWMutex保护共享数据。数据层MySQL存持久化数据，Redis存实时在线状态。"

//////////////////////////////////////

用户注册/登录
      │
      ▼
   MySQL ✅
   users表永久保存
   (重启服务器数据还在)

用户加入房间
      │
      ▼
   Redis ✅
   room:xxx:participants 实时更新
   (服务器重启数据消失，因为只需要当下状态)

GET /api/rooms 返回房间列表
      │
      ├── MySQL → 查rooms表拿房间基本信息
      │
      └── Redis → 查每个房间当前有几人
                  合并成 current_users 返回给前端