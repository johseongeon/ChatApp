package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"server/server_module"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type messageCollection struct {
	messageCol *mongo.Collection
	Mu         sync.Mutex
}

var Collection = &messageCollection{}

func getChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "room_id 쿼리 파라미터가 필요합니다", http.StatusBadRequest)
		return
	}

	filter := bson.M{"room_id": roomID}
	projection := bson.M{
		"_id":     0,
		"message": 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := Collection.messageCol.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		http.Error(w, "MongoDB 조회 실패", http.StatusInternalServerError)
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
	client, err := server_module.ConnectMongoDB()
	if err != nil {
		log.Fatal("MongoDB 연결 실패:", err)
	}

	// 컬렉션 지정
	Collection.messageCol = client.Database("ChatDB").Collection("messages")

	// HTTP 핸들러 등록
	http.HandleFunc("/history", getChatHistoryHandler)

	log.Println("chat_history_provider 서버 시작: :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("서버 실행 실패: %v", err)
	}
}
