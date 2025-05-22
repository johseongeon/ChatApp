package main

import (
	"context"
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
	opts := options.Update().SetUpsert(true) // 없으면 새로 만듦

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error adding friend:", err)
		return
	}
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

	// 예: /addFriend?username=A&friend=B
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

	fmt.Println("Server started on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
