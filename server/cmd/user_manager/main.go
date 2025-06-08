package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/pkg"
)

var UserManagerInstance = &pkg.UserManager{}

var Collection = &pkg.MessageCollection{}

func main() {
	client, err := pkg.ConnectMongoDB()
	if err != nil {
		log.Fatal("MongoDB 연결 실패:", err)
	}

	userManager := &pkg.UserManager{Client: client}
	RoomMgr := &pkg.RoomManager{Client: client}
	pkg.LoadRoomsFromDB(RoomMgr)

	// RoomManager 동기화
	go func() {
		for {
			pkg.LoadWhileRunning(RoomMgr)
			time.Sleep(3 * time.Second)
		}
	}()

	Collection.MessageCol = client.Database("ChatDB").Collection("users")

	//register
	http.HandleFunc("/register", pkg.RegisterServer(client))

	//addFriend
	http.HandleFunc("/addFriend", func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		friend := r.URL.Query().Get("friend")
		if username == "" || friend == "" {
			http.Error(w, "username and friend are required", http.StatusBadRequest)
			return
		}
		clientObj := &pkg.Client{Username: username}
		userManager.AddFriend(clientObj, friend)
		w.Write([]byte("Friend added successfully"))
	})

	//getFriends
	http.HandleFunc("/getFriends", func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &pkg.Client{Username: username}
		friends := userManager.GetFriends(clientObj)
		if friends == nil {
			http.Error(w, "Failed to get friends", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"friends":  friends,
		})
	})

	//getRooms
	http.HandleFunc("/getRooms", func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &pkg.Client{Username: username}
		rooms := userManager.GetChatRooms(clientObj)
		if rooms == nil {
			http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"rooms":    rooms,
		})
	})

	//createRoom
	http.HandleFunc("/createRoom", func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		roomID := r.URL.Query().Get("room_id")
		RoomMgr.CreateRoom(roomID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("created room successfully"))
	})

	//joinUser
	http.HandleFunc("/joinUser", func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		roomID := r.URL.Query().Get("room_id")
		clientObj := &pkg.Client{Username: username, Rooms: make(map[string]*pkg.ChatRoom)}
		room := RoomMgr.GetRoom(roomID)
		RoomMgr.JoinRoom(clientObj, room)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Joined room successfully"))
	})

	fmt.Println("Server started on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
