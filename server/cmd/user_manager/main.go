package main

import (
	"fmt"
	"log"
	"net/http"
	"server/pkg"
	"time"
)

var UserManagerInstance = &pkg.UserManager{}

var Collection = &pkg.MessageCollection{}

func main() {

	// connect to MongoDB
	client, err := pkg.ConnectMongoDB()
	if err != nil {
		log.Fatal("MongoDB 연결 실패:", err)
	}

	// Initialize UserManager and RoomManager
	userManager := &pkg.UserManager{Client: client}
	RoomMgr := &pkg.RoomManager{Client: client}
	// Load users from DB
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
	http.HandleFunc("/addFriend", pkg.Add_friend(client, userManager))

	//getFriends
	http.HandleFunc("/getFriends", pkg.GetFriends(client, userManager))

	//getRooms
	http.HandleFunc("/getRooms", pkg.GetRooms(client, userManager))

	//createRoom
	http.HandleFunc("/createRoom", pkg.CreateRoom(client, RoomMgr))

	//joinUser
	http.HandleFunc("/joinUser", pkg.JoinUser(client, RoomMgr))

	fmt.Println("Server started on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
