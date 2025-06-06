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

	pkg.LoadRoomsFromDB(RMG)

	c := &pkg.Client{
		Username: "B",
		Rooms:    make(map[string]*pkg.ChatRoom),
	}

	RMG.JoinRoom(c, RMG.GetRoom("test"))
}
