package main

import (
	"log"
	"net/http"
	handlers "server/http"
)

func main() {
	http.HandleFunc("/history", handlers.GetChatHistoryHandler)

	log.Println("chat_history_provider server start: :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
