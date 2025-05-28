# ChattApp

Go server for flutter chatting app

---

## real-time chatting server
- server/cmd/chat/main.go : websocket server for real-time chatting(create room, invite friend to room, and chatting)

start chat server first
``` cmd
cd server/cmd/chat
go run main.go
```

- client/main.dart : for CLI test 'server/chat/main.go'

run test (two or more cli)
``` cmd
cd client
dart run main.dart {username} {chatroom_id}
```

---
## chat history provider

- server/cmd/chat_history_provider/main.go : http server for providing chat history

start server first
```cmd
cd server/cmd/chat_history_provider
go run main.go
```

try
```cmd
curl http://localhost:8081/history?room_id={room_id}
```

---

## user_manager server
- server/cmd/user_manager : http server for register, managing friends and chatrooms

start server first

```cmd
cd server/cmd/user_manager
go run main.go
```

try

```cmd
curl http://localhost:8082/register?username={username}
curl "http://localhost:8082/addFriend?username={username}&friend={username}"
curl http://localhost:8082/getFriends?username={username}
curl http://localhost:8082/getRooms?username={username}
```
