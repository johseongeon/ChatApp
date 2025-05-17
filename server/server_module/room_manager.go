package server_module

import (
	"sync"
)

// handles all chat room operations
type RoomManager struct {
	Rooms map[string]*ChatRoom
	Mu    sync.RWMutex
}

var RoomMgr = &RoomManager{
	Rooms: make(map[string]*ChatRoom),
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
