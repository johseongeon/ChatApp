package main

import (
	"log"
	"net/http"
	"server/server_module"
	"time"

	"github.com/gorilla/websocket"
)

// upgrade to websocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	var initMsg struct {
		Username string `json:"username"`
		ChatID   string `json:"chat_id"`
	}
	log.Println("WebSocket connection attempt")
	err = conn.ReadJSON(&initMsg)
	if err != nil {
		log.Println("Failed to read init message:", err)
		return
	}

	client := &server_module.Client{
		Username: initMsg.Username,
		Conn:     conn,
		Rooms:    make(map[string]*server_module.ChatRoom),
	}

	room := server_module.RoomMgr.GetOrCreateRoom(initMsg.ChatID)
	client.JoinRoom(room)
	log.Printf("User %s joined chat %s", client.Username, room.Id)

	defer func() {
		for roomID := range client.Rooms {
			client.LeaveRoom(roomID)
		}
		conn.Close()
	}()

	for {
		var msg struct {
			Message string `json:"message"`
			RoomID  string `json:"room_id"`
		}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			return
		}

		roomID := initMsg.ChatID
		if msg.RoomID != "" {
			roomID = msg.RoomID
		}

		chatMsg := server_module.ChatMessage{
			Username:  client.Username,
			Message:   msg.Message,
			RoomID:    roomID,
			Timestamp: time.Now(),
		}
		if err := server_module.MessageLog.LogMessage(chatMsg); err != nil {
			log.Printf("Failed to log message: %v", err)
		}

		client.BroadcastToRoom(roomID, map[string]string{
			"from":    client.Username,
			"message": msg.Message,
		})
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
