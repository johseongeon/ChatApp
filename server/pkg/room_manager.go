package pkg

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// handles all chat room operations
type RoomManager struct {
	Rooms  map[string]*ChatRoom
	Mu     sync.RWMutex
	Client *mongo.Client
}

var RoomMgr = &RoomManager{
	Rooms:  make(map[string]*ChatRoom),
	Client: nil,
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

	// Update the MongoDB collection to add the room to the user's list
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := RoomMgr.Client.Database("ChatDB").Collection("users")

	filter := map[string]interface{}{"username": c.Username}
	update := map[string]interface{}{
		"$addToSet": map[string]interface{}{
			"rooms": room.Id,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error joining room :", err)
		return
	}
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
