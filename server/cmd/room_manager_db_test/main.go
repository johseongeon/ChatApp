package main

import (
	"log"
	"server/pkg"
)

func main() {

	mongoc, err := pkg.ConnectMongoDB()
	if err != nil {
		log.Fatal("Failed to connect MongoDB:", err)
	}

	RMG := &pkg.RoomManager{
		Client: mongoc,
		Rooms:  make(map[string]*pkg.ChatRoom),
	}

	c := &pkg.Client{
		Username: "testUser",
		Rooms:    make(map[string]*pkg.ChatRoom),
	}

	RMG.CreateRoom("c2")
	RMG.JoinRoom(c, RMG.GetRoom("c2"))

	room := RMG.GetRoom("c2")
	if room == nil {
		log.Fatal("Room c2 not found or not created properly")
	}
}
