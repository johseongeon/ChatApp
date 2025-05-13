package server_module

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Username string // identifier
	Conn     *websocket.Conn
	Rooms    map[string]*ChatRoom
	Mu       sync.RWMutex
}

type ChatRoom struct {
	Id      string // identifier
	Clients map[*Client]bool
	Mu      sync.RWMutex
}

// handles all chat room operations
type RoomManager struct {
	Rooms map[string]*ChatRoom
	Mu    sync.RWMutex
}

var RoomMgr = &RoomManager{
	Rooms: make(map[string]*ChatRoom),
}

type ChatMessage struct {
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	RoomID    string    `json:"room_id"`
	Timestamp time.Time `json:"timestamp"`
}

type MessageLogger struct {
	Mu sync.Mutex
}

var MessageLog = &MessageLogger{}

func (ml *MessageLogger) LogMessage(msg ChatMessage) error {
	ml.Mu.Lock()
	defer ml.Mu.Unlock()

	var logs []ChatMessage
	file, err := os.OpenFile("logs.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if stat, err := file.Stat(); err == nil && stat.Size() > 0 {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&logs); err != nil {
			return err
		}
	}

	logs = append(logs, msg)

	file.Seek(0, 0)
	file.Truncate(0)
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(logs)
}

func (rm *RoomManager) GetOrCreateRoom(roomID string) *ChatRoom {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	if room, exists := rm.Rooms[roomID]; exists {
		return room
	}

	room := &ChatRoom{
		Id:      roomID,
		Clients: make(map[*Client]bool),
	}
	rm.Rooms[roomID] = room
	return room
}

func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()
	delete(rm.Rooms, roomID)
}

func (c *Client) JoinRoom(room *ChatRoom) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	room.Mu.Lock()
	defer room.Mu.Unlock()

	c.Rooms[room.Id] = room
	room.Clients[c] = true
}

func (c *Client) LeaveRoom(roomID string) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if room, exists := c.Rooms[roomID]; exists {
		room.Mu.Lock()
		delete(room.Clients, c)
		room.Mu.Unlock()

		delete(c.Rooms, roomID)

		room.Mu.RLock()
		if len(room.Clients) == 0 {
			room.Mu.RUnlock()
			RoomMgr.RemoveRoom(roomID)
		} else {
			room.Mu.RUnlock()
		}
	}
}

func (c *Client) BroadcastToRoom(roomID string, message map[string]string) {
	c.Mu.RLock()
	room, exists := c.Rooms[roomID]
	c.Mu.RUnlock()

	if !exists {
		return
	}

	room.Mu.RLock()
	defer room.Mu.RUnlock()

	for client := range room.Clients {
		if client != c {
			client.Conn.WriteJSON(message)
		}
	}
}
