package handlers

import (
	"encoding/json"
	"net/http"
	"server/pkg"

	"go.mongodb.org/mongo-driver/mongo"
)

func CreateRoom(client *mongo.Client, rm *pkg.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		roomID := r.URL.Query().Get("room_id")
		rm.CreateRoom(roomID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("created room successfully"))
	}
}

func GetRooms(client *mongo.Client, userManager *pkg.UserManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &pkg.Client{Username: username}
		rooms := userManager.GetRooms(clientObj)
		if rooms == nil {
			http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"rooms":    rooms,
		})
	}
}

func JoinUser(client *mongo.Client, rm *pkg.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		roomID := r.URL.Query().Get("room_id")
		clientObj := &pkg.Client{Username: username, Rooms: make(map[string]*pkg.ChatRoom)}
		room := rm.GetRoom(roomID)
		rm.JoinRoom(clientObj, room)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Joined room successfully"))
	}
}
