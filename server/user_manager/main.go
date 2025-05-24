package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/server_module"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserManager struct {
	Mu     sync.Mutex
	Client *mongo.Client
}

var UserManagerInstance = &UserManager{}

var Collection = &server_module.MessageCollection{}

func RegisterUser(client *mongo.Client, username string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("ChatDB").Collection("users")

	// 중복 방지
	filter := map[string]interface{}{"username": username}
	update := map[string]interface{}{
		"$setOnInsert": map[string]interface{}{
			"username": username,
			"friends":  []string{},
			"rooms":    []string{},
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error registering user:", err)
		return
	}

	fmt.Println("User registered:", username)
}

func (adder *UserManager) AddFriend(c *server_module.Client, friend string) {
	adder.Mu.Lock()
	defer adder.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := adder.Client.Database("ChatDB").Collection("users")

	filter := map[string]interface{}{"username": c.Username}
	update := map[string]interface{}{
		"$addToSet": map[string]interface{}{
			"friends": friend,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error adding friend:", err)
		return
	}
}

func (adder *UserManager) RemoveFriend(c *server_module.Client, friend string) {} // 추가 필요

func (adder *UserManager) GetFriends(c *server_module.Client) []string {
	adder.Mu.Lock()
	defer adder.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := adder.Client.Database("ChatDB").Collection("users")

	filter := map[string]interface{}{"username": c.Username}
	projection := map[string]interface{}{
		"friends": 1,
	}

	var result struct {
		Friends []string `bson:"friends"`
	}

	err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		log.Println("Error getting friends:", err)
		return nil
	}

	return result.Friends
}

func (adder *UserManager) GetChatRooms(c *server_module.Client) []string {
	adder.Mu.Lock()
	defer adder.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := adder.Client.Database("ChatDB").Collection("users")

	filter := map[string]interface{}{"username": c.Username}
	projection := map[string]interface{}{
		"rooms": 1,
	}

	var result struct {
		Rooms []string `bson:"rooms"`
	}

	err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		log.Println("Error getting rooms:", err)
		return nil
	}

	return result.Rooms
}

func RegisterServer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		RegisterUser(client, username)
		w.Write([]byte("User registered successfully"))
	}
}

// curl http://localhost:8082/register?username=A
// curl "http://localhost:8082/addFriend?username=A&friend=B"
func main() {
	client, err := server_module.ConnectMongoDB()
	if err != nil {
		log.Fatal("MongoDB 연결 실패:", err)
	}

	userManager := &UserManager{Client: client}
	Collection.MessageCol = client.Database("ChatDB").Collection("users")

	http.HandleFunc("/register", RegisterServer(client))

	http.HandleFunc("/addFriend", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		friend := r.URL.Query().Get("friend")
		if username == "" || friend == "" {
			http.Error(w, "username and friend are required", http.StatusBadRequest)
			return
		}
		clientObj := &server_module.Client{Username: username}
		userManager.AddFriend(clientObj, friend)
		w.Write([]byte("Friend added successfully"))
	})

	http.HandleFunc("/getFriends", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &server_module.Client{Username: username}
		friends := userManager.GetFriends(clientObj)
		if friends == nil {
			http.Error(w, "Failed to get friends", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"friends":  friends,
		})
	})

	http.HandleFunc("/getRooms", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &server_module.Client{Username: username}
		rooms := userManager.GetChatRooms(clientObj)
		if rooms == nil {
			http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"rooms":    rooms,
		})
	})

	fmt.Println("Server started on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
