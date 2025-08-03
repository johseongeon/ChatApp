package main

import (
	"log"
	"net/http"
	"server/db"
	handlers "server/http"
	"server/pkg"
	"time"
)

// upgrade to websocket

func main() {
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	log.Println("Server started on :8080")
	client, err := db.ConnectMongoDB()
	if err != nil {
		log.Fatal("Failed to connect MongoDB:", err)
	}
	pkg.MessageLog.Client = client
	pkg.RoomMgr.Client = client
	db.LoadRoomsFromDB(pkg.RoomMgr)

	// RoomManager 동기화
	go func() {
		for {
			db.LoadWhileRunning(pkg.RoomMgr)
			time.Sleep(3 * time.Second)
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
