package db

import (
	"context"
	"log"
	"os"
	"server/pkg"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageCollection struct {
	MessageCol *mongo.Collection
	Mu         sync.Mutex
}

func ConnectMongoDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func LoadWhileRunning(mgr *pkg.RoomManager) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	collection := mgr.Client.Database("ChatDB").Collection("rooms")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error loading rooms from DB: %v", err)
		return
	}
	defer cursor.Close(ctx)

	mgr.Mu.Lock()
	defer mgr.Mu.Unlock()

	for cursor.Next(ctx) {
		var roomDoc struct {
			RoomID  string   `bson:"room_id"`
			Clients []string `bson:"clients"`
		}
		if err := cursor.Decode(&roomDoc); err != nil {
			log.Printf("Error decoding room document: %v", err)
			continue
		}

		room, exists := mgr.Rooms[roomDoc.RoomID]
		if !exists {
			// 기존에 없던 새 room만 생성
			room = &pkg.ChatRoom{
				Id:      roomDoc.RoomID,
				Clients: make(map[*pkg.Client]bool),
			}
			mgr.Rooms[roomDoc.RoomID] = room
		}
	}
}

func LoadRoomsFromDB(mgr *pkg.RoomManager) {
	mgr.Mu.Lock()
	defer mgr.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := mgr.Client.Database("ChatDB").Collection("rooms")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error loading rooms from DB: %v", err)
		return
	}
	defer cursor.Close(ctx)

	mgr.Rooms = make(map[string]*pkg.ChatRoom) // initialize the map

	for cursor.Next(ctx) {
		var roomDoc struct {
			RoomID  string   `bson:"room_id"`
			Clients []string `bson:"clients"`
		}

		if err := cursor.Decode(&roomDoc); err != nil {
			log.Printf("Error decoding room document: %v", err)
			continue
		}

		mgr.Rooms[roomDoc.RoomID] = &pkg.ChatRoom{
			Id:      roomDoc.RoomID,
			Clients: make(map[*pkg.Client]bool),
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error after iteration: %v", err)
	}
}
