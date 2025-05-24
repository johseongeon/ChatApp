package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"server/pkg"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Collection = &pkg.MessageCollection{}

func getChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "room_id query parameter requried.", http.StatusBadRequest)
		return
	}

	filter := bson.M{"room_id": roomID}
	projection := bson.M{
		"_id":     0,
		"message": 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := Collection.MessageCol.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		http.Error(w, "Failed to Find MongoDB", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		http.Error(w, "데이터 파싱 실패", http.StatusInternalServerError)
		return
	}

	messages := make([]string, 0, len(results))
	for _, doc := range results {
		if msg, ok := doc["message"].(string); ok {
			messages = append(messages, msg)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func main() {
	// MongoDB 연결
	client, err := pkg.ConnectMongoDB()
	if err != nil {
		log.Fatal("Failed to connect MongoDB:", err)
	}

	// 컬렉션 지정
	Collection.MessageCol = client.Database("ChatDB").Collection("messages")

	// HTTP 핸들러 등록
	http.HandleFunc("/history", getChatHistoryHandler)

	log.Println("chat_history_provider server start: :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
