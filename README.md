# ChattApp

Go server for flutter chatting app

---
## server
- server/cmd/chat/main.go : websocket server for real-time chatting(create room, invite friend to room, and chatting)
- server/cmd/chat_history_provider/main.go : http server for providing chat history
- server/cmd/user_manager : http server for register, managing friends and chatrooms
---

- client/main.dart : for CLI test 'server/chat/main.go'

start chat server first
``` cmd
cd server/cmd/chat
go run main.go
```

run test (two or more cli)
``` cmd
cd client
dart run main.dart {username} {chatroom_id}
```
