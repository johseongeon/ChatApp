# ChattApp

![MongoDB](https://img.shields.io/badge/MongoDB-%234ea94b.svg?style=for-the-badge&logo=mongodb&logoColor=white)
![Dart](https://img.shields.io/badge/dart-%230175C2.svg?style=for-the-badge&logo=dart&logoColor=white)
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)

Go server for flutter chatting app

Make sure that MongoDB is installed on your local system, and mongo server is listening at port 27017

---

## Using Guide

1. Register User First
2. Add friend, Get Friend List
3. Create Chatroom and Invite/Join
4. Real-time Chatting

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
//register and add friend
curl http://localhost:8082/register?username={username}
curl "http://localhost:8082/addFriend?username={username}&friend={username}"

//create room and invite user
curl http://localhost:8082/createRoom?room_id={room_id}
curl "http://localhost:8082/joinUser?room_id={room_id}&username={username}"

//get list of friends and rooms
curl http://localhost:8082/getFriends?username={username}
curl http://localhost:8082/getRooms?username={username}
```

---

## real-time chatting server
- server/cmd/chat/main.go : websocket server for real-time chatting

- You must create room and invite users to the room before test.
- If not, it returns a nil value, which causes an error.

start chat server first
``` cmd
cd server/cmd/chat
go run main.go
```

- client/main.dart : for CLI test 'server/chat/main.go'

run test via dart cli (two or more)
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

# Run Server Using Docker Compose

If docker is installed, you can use docker-compose to run server

``` docker-compose.yml
version: '3.8'

services:
  mongo:
    image: mongo:6.0
    container_name: chat-mongo
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

  cluster-manager:
    image: johseongeon/user-manager:1.0
    container_name: user-manager
    ports:
      - "8082:8082"
    environment:
      - MONGO_URL=mongodb://mongo:27017
    depends_on:
      - mongo

  chat-history-provider:
    image: johseongeon/chat-history-provider:1.0
    container_name: chat-history-provider
    ports:
      - "8081:8081"
    environment:
      - MONGO_URL=mongodb://mongo:27017
    depends_on:
      - mongo

  chat-server:
    image: johseongeon/chat-server:1.0
    container_name: chat-server
    ports:
      - "8080:8080"
    environment:
      - MONGO_URL=mongodb://mongo:27017
    depends_on:
      - mongo

volumes:
  mongo_data:
```

copy this docker-compose.yml and run

```cmd
docker compose up --build
```

Then you don't have to run 3 servers one-by-one
