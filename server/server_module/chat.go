package server_module

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}
	return client, nil
}

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
	Username  string    `json:"username" bson:"username"`
	Message   string    `json:"message" bson:"message"`
	RoomID    string    `json:"room_id" bson:"room_id"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type MessageLogger struct {
	Mu     sync.Mutex
	Client *mongo.Client
}

var MessageLog = &MessageLogger{}

func (ml *MessageLogger) LogMessage(msg ChatMessage) error {
	ml.Mu.Lock()
	defer ml.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := ml.Client.
		Database("ChatDB").
		Collection("messages")

	_, err := collection.InsertOne(ctx, msg)
	return err
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
