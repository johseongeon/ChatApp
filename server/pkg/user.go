package pkg

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserManager struct {
	Mu     sync.Mutex
	Client *mongo.Client
}

func RegisterUser(client *mongo.Client, username string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("ChatDB").Collection("users")

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

func (adder *UserManager) AddFriend(c *Client, friend string) {
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

func (adder *UserManager) GetFriends(c *Client) []string {
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
