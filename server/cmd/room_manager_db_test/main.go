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

	RMG.CreateRoom("t2")
	RMG.RemoveRoom("t1")
}
