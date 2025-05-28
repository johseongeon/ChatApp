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
	pkg.EnableCORS(w)
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "room_id query parameter required.", http.StatusBadRequest)
		return
	}

	filter := bson.M{"room_id": roomID}
	projection := bson.M{
		"_id":       0,
		"username":  1,
		"message":   1,
		"timestamp": 1,
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
		http.Error(w, "Failed to parse data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	client, err := pkg.ConnectMongoDB()
	if err != nil {
		log.Fatal("Failed to connect MongoDB:", err)
	}

	Collection.MessageCol = client.Database("ChatDB").Collection("messages")

	http.HandleFunc("/history", getChatHistoryHandler)

	log.Println("chat_history_provider server start: :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
