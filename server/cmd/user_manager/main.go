package main

import (
	"fmt"
	"log"
	"net/http"
	"server/db"
	handlers "server/http"
	"server/pkg"
	"time"
)

var UserManagerInstance = &pkg.UserManager{}

var Collection = &db.MessageCollection{}

func main() {

	// connect to MongoDB
	client, err := db.ConnectMongoDB()
	if err != nil {
		log.Fatal("MongoDB 연결 실패:", err)
	}

	// Initialize UserManager and RoomManager
	userManager := &pkg.UserManager{Client: client}
	RoomMgr := &pkg.RoomManager{Client: client}
	// Load users from DB
	db.LoadRoomsFromDB(RoomMgr)

	// RoomManager 동기화
	go func() {
		for {
			db.LoadWhileRunning(RoomMgr)
			time.Sleep(3 * time.Second)
		}
	}()

	Collection.MessageCol = client.Database("ChatDB").Collection("users")

	//register
	http.HandleFunc("/register", handlers.RegisterServer(client))

	//addFriend
	http.HandleFunc("/addFriend", handlers.Add_friend(client, userManager))

	//getFriends
	http.HandleFunc("/getFriends", handlers.GetFriends(client, userManager))

	//getRooms
	http.HandleFunc("/getRooms", handlers.GetRooms(client, userManager))

	//createRoom
	http.HandleFunc("/createRoom", handlers.CreateRoom(client, RoomMgr))

	//joinUser
	http.HandleFunc("/joinUser", handlers.JoinUser(client, RoomMgr))

	fmt.Println("Server started on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
