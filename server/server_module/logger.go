package server_module

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type MessageLogger struct {
	Mu     sync.Mutex
	Client *mongo.Client
}

var MessageLog = &MessageLogger{}

func (ml *MessageLogger) LogMessage(msg ChatMessage) error {
	ml.Mu.Lock()
	defer ml.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := ml.Client.
		Database("ChatDB").
		Collection("messages")

	_, err := collection.InsertOne(ctx, msg)
	return err
}
