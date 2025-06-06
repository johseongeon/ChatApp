package pkg

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

func (rm *RoomManager) GetRoom(roomID string) *ChatRoom {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	if room, exists := rm.Rooms[roomID]; exists {
		return room
	}

	return nil
}

func LoadRoomsFromDB(mgr *RoomManager) {
	mgr.Mu.Lock()
	defer mgr.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := mgr.Client.Database("ChatDB").Collection("rooms")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error loading rooms from DB: %v", err)
		return
	}
	defer cursor.Close(ctx)

	mgr.Rooms = make(map[string]*ChatRoom) // initialize the map

	for cursor.Next(ctx) {
		var roomDoc struct {
			RoomID  string   `bson:"room_id"`
			Clients []string `bson:"clients"` // 클라이언트 목록은 현재 필요 없지만 추후 확장 가능
		}

		if err := cursor.Decode(&roomDoc); err != nil {
			log.Printf("Error decoding room document: %v", err)
			continue
		}

		mgr.Rooms[roomDoc.RoomID] = &ChatRoom{
			Id:      roomDoc.RoomID,
			Clients: make(map[*Client]bool),
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error after iteration: %v", err)
	}
}

func (rm *RoomManager) ConnectToRoom(client *Client, room *ChatRoom) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	if room == nil {
		log.Println("Room is nil, cannot connect.")
		return
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	if _, exists := room.Clients[client]; !exists {
		room.Clients[client] = true
		client.Rooms[room.Id] = room

		log.Printf("Client %s connected to room %s", client.Username, room.Id)
	} else {
		log.Printf("Client %s is already connected to room %s", client.Username, room.Id)
	}
}

func (rm *RoomManager) CreateRoom(roomID string) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	room := &ChatRoom{
		Id:      roomID,
		Clients: make(map[*Client]bool),
	}
	rm.Rooms[roomID] = room

	// Update the MongoDB collection to add the room
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := rm.Client.Database("ChatDB").Collection("rooms")
	filter := map[string]interface{}{"room_id": roomID}
	update := map[string]interface{}{
		"$setOnInsert": map[string]interface{}{
			"room_id": roomID,
			"clients": []string{},
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error creating room:", err)
		return
	}
}

func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()
	delete(rm.Rooms, roomID)

	// Update the MongoDB collection to remove the room
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := rm.Client.Database("ChatDB").Collection("rooms")
	filter := bson.M{"room_id": roomID}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println("Error removing room:", err)
		return
	}

}

func (rm *RoomManager) JoinRoom(c *Client, room *ChatRoom) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	room.Mu.Lock()
	defer room.Mu.Unlock()

	c.Rooms[room.Id] = room
	room.Clients[c] = true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := rm.Client.Database("ChatDB").Collection("users")

	userFilter := map[string]interface{}{"username": c.Username}
	userUpdate := map[string]interface{}{
		"$addToSet": map[string]interface{}{
			"rooms": room.Id,
		},
	}

	_, err := userCollection.UpdateOne(ctx, userFilter, userUpdate)
	if err != nil {
		log.Println("Error updating user's rooms in 'users' collection:", err)
		return
	}

	roomCollection := rm.Client.Database("ChatDB").Collection("rooms")

	roomFilter := map[string]interface{}{"room_id": room.Id}
	roomUpdate := map[string]interface{}{
		"$addToSet": map[string]interface{}{
			"clients": c.Username,
		},
	}

	_, err = roomCollection.UpdateOne(ctx, roomFilter, roomUpdate)
	if err != nil {
		log.Println("Error updating room's clients in 'rooms' collection:", err)
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
